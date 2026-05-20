# frozen_string_literal: true

require 'json'
require 'osvm'
require 'shellwords'
require 'test-runner/hook'

class Vpsadminctl
  def initialize(machine)
    @machine = machine
  end

  def succeeds(args:, opts: {}, parameters: {}, timeout: nil)
    run_with_timeout(:succeeds, args:, opts:, parameters:, timeout:)
  end

  private

  def run_with_timeout(method, args:, opts:, parameters:, timeout: nil)
    cmd = build_command(args:, opts:, parameters:)
    status, output =
      if timeout
        @machine.public_send(method, cmd, timeout:)
      else
        @machine.public_send(method, cmd)
      end

    [status, parse_output(output)]
  end

  def build_command(args:, opts:, parameters:)
    cmd = ['vpsadminctl', '--raw']
    cmd.concat(format_options(opts))
    cmd.concat(args.map(&:to_s))

    unless parameters.empty?
      cmd << '--'
      cmd.concat(format_options(parameters))
    end

    Shellwords.join(cmd)
  end

  def format_options(options)
    options.flat_map do |key, value|
      name = key.to_s.tr('_', '-')
      flag = name.length == 1 ? "-#{name}" : "--#{name}"

      case value
      when true
        [flag]
      when false
        [flag.sub(/^--/, '--no-')]
      when nil
        []
      when Array
        value.flat_map { |v| [flag, v.to_s] }
      else
        [flag, value.to_s]
      end
    end
  end

  def parse_output(output)
    json = extract_json_prefix(output)
    return output if json.nil?

    JSON.parse(json)
  rescue JSON::ParserError
    output
  end

  def extract_json_prefix(text)
    return nil unless text

    start = text.index(/\{|\[/)
    return nil if start.nil?

    stack = []
    in_string = false
    escape = false

    text.chars.each_with_index do |ch, idx|
      next if idx < start

      if in_string
        if escape
          escape = false
        elsif ch == '\\'
          escape = true
        elsif ch == '"'
          in_string = false
        end

        next
      end

      case ch
      when '"'
        in_string = true
      when '{'
        stack << '}'
      when '['
        stack << ']'
      when '}', ']'
        return nil if stack.empty?

        expected = stack.pop
        return nil if ch != expected

        return text[start..idx] if stack.empty?
      end
    end

    nil
  end
end

class VpsadminServicesMachine < OsVm::NixosMachine
  CHAIN_STATES = {
    staged: 0,
    queued: 1,
    done: 2,
    rollbacking: 3,
    failed: 4,
    fatal: 5,
    resolved: 6
  }.freeze

  attr_reader :vpsadminctl

  def initialize(...)
    super
    @vpsadminctl = Vpsadminctl.new(self)
  end

  def wait_for_vpsadmin_api(timeout: @default_timeout || 300)
    deadline = Time.now + timeout

    loop do
      raise OsVm::TimeoutError, 'Timed out waiting for vpsAdmin API' if Time.now >= deadline

      _, output = wait_until_succeeds(
        'curl --silent --fail-with-body http://api.vpsadmin.test/',
        timeout: [1, (deadline - Time.now).ceil].max
      )

      return true if output.include?('API description')

      sleep 1
    end
  end

  def api_ruby(code:, timeout: nil)
    script = <<~CMD
      set -euo pipefail
      api_dir="$(systemctl show -p WorkingDirectory --value vpsadmin-api)"
      api_root="$(dirname "$api_dir")"
      tmp_rb="$(mktemp /tmp/vpsadmin-provider-it-XXXX.rb)"
      trap 'rm -f "$tmp_rb"' EXIT

      cat > "$tmp_rb" <<'RUBY'
      ENV['RACK_ENV'] ||= 'production'
      require 'json'
      Dir.chdir(ENV.fetch('API_DIR'))
      $LOAD_PATH.unshift(File.join(ENV.fetch('API_DIR'), 'lib'))
      require 'vpsadmin'
      #{code}
      RUBY

      API_DIR="$api_dir" "$api_root/ruby-env-wrapped/bin/ruby" "$tmp_rb"
    CMD

    timeout ? succeeds(script, timeout:) : succeeds(script)
  end

  def api_ruby_json(code:, timeout: nil)
    _, output = api_ruby(code:, timeout:)
    JSON.parse(output.to_s.lines.last)
  end

  def mariadb_raw(sql:, database: 'vpsadmin', user: 'api', timeout: nil)
    cmd = mariadb_command(sql:, database:, user:)
    timeout ? succeeds(cmd, timeout:) : succeeds(cmd)
  end

  def mariadb_scalar(sql:, database: 'vpsadmin', user: 'api', timeout: nil)
    _, output = mariadb_raw(sql:, database:, user:, timeout:)
    output.to_s.lines.first&.strip
  end

  def wait_for_chain_state(chain_id, state:, timeout: @default_timeout || 300)
    expected = state.is_a?(Symbol) ? CHAIN_STATES.fetch(state) : Integer(state)

    wait_for_condition(
      timeout:,
      error_message: "Timed out waiting for chain ##{chain_id} state=#{state}"
    ) do
      current = mariadb_scalar(
        sql: "SELECT state FROM transaction_chains WHERE id = #{Integer(chain_id)}"
      )

      current && current.to_i == expected
    end
  end

  def unlock_transaction_signing_key(passphrase: 'test')
    vpsadminctl.succeeds(
      args: %w[api_server unlock_transaction_signing_key],
      parameters: { passphrase: }
    )
  end

  private

  def mariadb_command(sql:, database:, user:)
    password_file = "/etc/vpsadmin-test/mariadb-#{user}-password"

    inner = [
      'mariadb',
      '--batch',
      '--raw',
      '--skip-column-names',
      "--user=#{Shellwords.escape(user)}",
      "--password=\"$(cat #{Shellwords.escape(password_file)})\"",
      Shellwords.escape(database),
      '-e',
      Shellwords.escape(sql)
    ].join(' ')

    "bash -lc #{Shellwords.escape(inner)}"
  end

  def wait_for_condition(timeout:, error_message:)
    deadline = Time.now + timeout

    loop do
      return true if yield

      raise OsVm::TimeoutError, error_message if Time.now >= deadline

      sleep 1
    end
  end
end

TestRunner::Hook.subscribe(:machine_class_for) do |machine_config|
  next unless machine_config.tags.include?('vpsadmin-services')

  VpsadminServicesMachine
end
