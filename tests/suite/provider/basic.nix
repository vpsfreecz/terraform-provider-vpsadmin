{
  vpsadminPath,
  providerPackage,
  ...
}@args:
let
  seed = import (vpsadminPath + "/api/db/seeds/test-1-node.nix");
  adminUser = seed.adminUser;
  location = seed.location;
  nodeId = seed.node.id;
  common = import (vpsadminPath + "/tests/suite/storage/remote-common.nix") {
    adminUserId = adminUser.id;
    node1Id = nodeId;
    node2Id = nodeId;
    manageCluster = false;
  };
  extraModules = args.extraModules or { };
  clusterArgs = args // {
    extraModules = extraModules // {
      services =
        { pkgs, ... }:
        {
          imports = if extraModules ? services then [ extraModules.services ] else [ ];

          environment.systemPackages = [
            pkgs.opentofu
            providerPackage
          ];
        };
    };
  };
in
import ../../make-test.nix (
  { ... }:
  {
    name = "provider-basic";

    description = ''
      Exercise the Terraform provider against a local single-node vpsAdmin
      cluster using OpenTofu and provider development overrides.
    '';

    tags = [
      "ci"
      "provider"
      "vpsadmin"
    ];

    machines = import (vpsadminPath + "/tests/machines/cluster/1-node.nix") clusterArgs;

    testScript = common + ''
      configure_examples do |config|
        config.default_order = :defined
      end

      workdir = '/tmp/provider-basic'
      api_url = 'http://api.vpsadmin.test/'
      location = ${builtins.toJSON location.label}
      os_template = 'debian-latest-x86_64-vpsadminos-minimal'
      public_key = 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBKuiydHSKEYK5QbvSOgRZpp4mmhUSr9eg9HiTjYpUrY provider-basic@example.test'

      def tofu_env(workdir, api_url, token)
        {
          'TF_CLI_CONFIG_FILE' => File.join(workdir, 'tofurc'),
          'TF_VAR_api_url' => api_url,
          'TF_VAR_auth_token' => token
        }
      end

      def env_prefix(env)
        env.map { |k, v| "#{k}=#{Shellwords.escape(v)}" }.join(' ')
      end

      def tofu(services, workdir, api_url, token, args, timeout: 1800)
        cmd = [
          env_prefix(tofu_env(workdir, api_url, token)),
          'tofu',
          "-chdir=#{Shellwords.escape(workdir)}",
          args
        ].join(' ')

        services.succeeds(cmd, timeout: timeout)
      end

      def tofu_outputs(services, workdir, api_url, token)
        _, output = tofu(services, workdir, api_url, token, 'output -json 2>/dev/null', timeout: 300)
        JSON.parse(output)
      end

      def expect_output(outputs, name, expected)
        expect(outputs.fetch(name).fetch('value')).to eq(expected)
      end

      def write_file(machine, path, content)
        machine.succeeds("install -d #{Shellwords.escape(File.dirname(path))}")
        machine.succeeds(<<~CMD)
          cat > #{Shellwords.escape(path)} <<'EOF'
          #{content}
          EOF
        CMD
      end

      def write_tofu_cli_config(machine, workdir)
        write_file(machine, File.join(workdir, 'tofurc'), <<~HCL)
          provider_installation {
            dev_overrides {
              "vpsfreecz/vpsadmin" = "/run/current-system/sw/bin"
            }
            direct {}
          }
        HCL
      end

      def write_provider_config(machine, workdir, public_key, location, os_template, attrs)
        write_file(machine, File.join(workdir, 'main.tf'), <<~HCL)
          terraform {
            required_providers {
              vpsadmin = {
                source = "vpsfreecz/vpsadmin"
              }
            }
          }

          variable "api_url" {
            type = string
          }

          variable "auth_token" {
            type      = string
            sensitive = true
          }

          provider "vpsadmin" {
            api_url    = var.api_url
            auth_token = var.auth_token
          }

          resource "vpsadmin_ssh_key" "provider_it" {
            label    = "provider-it"
            key      = #{public_key.inspect}
            auto_add = #{attrs.fetch(:ssh_key_auto_add)}
          }

          data "vpsadmin_ssh_key" "provider_it" {
            label = vpsadmin_ssh_key.provider_it.label

            depends_on = [vpsadmin_ssh_key.provider_it]
          }

          resource "vpsadmin_vps" "provider_it" {
            location            = #{location.inspect}
            install_os_template = #{os_template.inspect}
            hostname            = "provider-it"
            cpu                 = 1
            memory              = #{attrs.fetch(:vps_memory)}
            swap                = 0
            diskspace           = 4096
            public_ipv4_count   = 0
            private_ipv4_count  = 0
            public_ipv6_count   = 0
            feature_fuse        = false
            feature_kvm         = false
            feature_lxc         = false
            feature_ppp         = false
            feature_tun         = false
          }

          data "vpsadmin_vps" "provider_it" {
            vps_id = vpsadmin_vps.provider_it.id

            depends_on = [vpsadmin_vps.provider_it]
          }

          resource "vpsadmin_dataset" "provider_it" {
            name     = "vps''${vpsadmin_vps.provider_it.id}/provider-it-subdataset"
            refquota = #{attrs.fetch(:dataset_refquota)}
          }

          data "vpsadmin_dataset" "provider_it" {
            name = vpsadmin_dataset.provider_it.name

            depends_on = [vpsadmin_dataset.provider_it]
          }

          resource "vpsadmin_mount" "provider_it" {
            vps           = vpsadmin_vps.provider_it.id
            dataset       = vpsadmin_dataset.provider_it.id
            mountpoint    = "/mnt/provider-it-subdataset"
            enable        = #{attrs.fetch(:mount_enable)}
            mode          = "rw"
            on_start_fail = "mount_later"
          }

          data "vpsadmin_mount" "provider_it" {
            vps      = vpsadmin_vps.provider_it.id
            mount_id = vpsadmin_mount.provider_it.id

            depends_on = [vpsadmin_mount.provider_it]
          }

          output "vps_id" {
            value = vpsadmin_vps.provider_it.id
          }

          output "ssh_key_auto_add" {
            value = data.vpsadmin_ssh_key.provider_it.auto_add
          }

          output "vps_memory" {
            value = data.vpsadmin_vps.provider_it.memory
          }

          output "dataset_refquota" {
            value = data.vpsadmin_dataset.provider_it.refquota
          }

          output "mount_enable" {
            value = data.vpsadmin_mount.provider_it.enable
          }
        HCL
      end

      def create_provider_token(services, admin_user_id)
        services.api_ruby_json(code: <<~RUBY)
          user = User.find(#{admin_user_id})
          user.update!(level: 2)
          user_agent = UserAgent.find_or_create!('provider-basic integration test')

          session = Token.for_new_record!(Time.now + 3600) do |token|
            UserSession.create!(
              user: user,
              admin: nil,
              user_agent: user_agent,
              auth_type: 'token',
              api_ip_addr: '127.0.0.1',
              api_ip_ptr: 'localhost',
              client_ip_addr: '127.0.0.1',
              client_ip_ptr: 'localhost',
              client_version: 'terraform-provider-vpsadmin integration test',
              token: token,
              token_str: token.token,
              token_lifetime: 'renewable_manual',
              token_interval: 3600,
              scope: ['all'],
              label: 'provider-basic'
            )
          end

          puts JSON.dump(token: session.token.token, session_id: session.id)
        RUBY
      end

      def configure_vps_soft_delete_lifetime(services, location)
        services.api_ruby_json(code: <<~RUBY)
          env = Location.find_by!(label: #{location.inspect}).environment
          lifetime = DefaultLifetimeValue.find_or_initialize_by(
            environment: env,
            class_name: 'Vps',
            direction: DefaultLifetimeValue.directions.fetch('enter'),
            state: DefaultLifetimeValue.states.fetch('soft_delete')
          )
          lifetime.reason = 'provider-basic integration test cleanup'
          lifetime.add_expiration = 3600
          lifetime.save!

          puts JSON.dump(default_lifetime_value_id: lifetime.id)
        RUBY
      end

      before(:suite) do
        [services, node].each(&:start)
        services.wait_for_vpsadmin_api
        wait_for_running_nodectld(node)
        wait_for_node_ready(services, node1_id)
        services.unlock_transaction_signing_key(passphrase: 'test')
      end

      after(:suite) do
        if @token
          services.execute(
            "#{env_prefix(tofu_env(workdir, api_url, @token))} tofu -chdir=#{Shellwords.escape(workdir)} destroy -auto-approve -input=false",
            timeout: 1800
          )
        end
      end

      describe 'provider basic', order: :defined do
        it 'prepares vpsAdmin and OpenTofu inputs' do
          pool = create_pool(
            services,
            node_id: node1_id,
            label: 'provider-basic',
            filesystem: primary_pool_fs,
            role: 'hypervisor'
          )
          wait_for_pool_online(services, pool.fetch('id'))

          configure_vps_soft_delete_lifetime(services, location)
          @token = create_provider_token(services, admin_user_id).fetch('token')
          services.succeeds("rm -rf #{Shellwords.escape(workdir)} && install -d #{Shellwords.escape(workdir)}")
          write_tofu_cli_config(services, workdir)
          write_provider_config(
            services,
            workdir,
            public_key,
            location,
            os_template,
            {
              ssh_key_auto_add: false,
              vps_memory: 1024,
              dataset_refquota: 1024,
              mount_enable: false
            }
          )
        end

        it 'creates provider resources and reads data sources' do
          tofu(services, workdir, api_url, @token, 'apply -auto-approve -input=false')

          outputs = tofu_outputs(services, workdir, api_url, @token)
          @vps_id = outputs.fetch('vps_id').fetch('value')
          expect(@vps_id).to match(/\A[0-9]+\z/)
          expect_output(outputs, 'ssh_key_auto_add', false)
          expect_output(outputs, 'vps_memory', 1024)
          expect_output(outputs, 'dataset_refquota', 1024)
          expect_output(outputs, 'mount_enable', false)
        end

        it 'updates resources and converges without further changes' do
          write_provider_config(
            services,
            workdir,
            public_key,
            location,
            os_template,
            {
              ssh_key_auto_add: true,
              vps_memory: 1536,
              dataset_refquota: 2048,
              mount_enable: true
            }
          )

          tofu(services, workdir, api_url, @token, 'apply -auto-approve -input=false')

          outputs = tofu_outputs(services, workdir, api_url, @token)
          expect_output(outputs, 'ssh_key_auto_add', true)
          expect_output(outputs, 'vps_memory', 1536)
          expect_output(outputs, 'dataset_refquota', 2048)
          expect_output(outputs, 'mount_enable', true)

          tofu(services, workdir, api_url, @token, 'plan -detailed-exitcode -input=false', timeout: 600)
        end

        it 'destroys provider resources' do
          services.vpsadminctl.succeeds(
            args: ['vps', 'stop', @vps_id],
            parameters: { force: true }
          )
          wait_for_vps_on_node(services, vps_id: @vps_id, node_id: node1_id, running: false)

          tofu(services, workdir, api_url, @token, 'destroy -auto-approve -input=false')
          @token = nil
        end
      end
    '';
  }
) clusterArgs
