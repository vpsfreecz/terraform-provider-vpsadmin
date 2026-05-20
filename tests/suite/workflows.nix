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
import ../make-test.nix (
  { ... }:
  {
    name = "workflows";

    description = ''
      Exercise user workflows for the Terraform provider against a local
      single-node vpsAdmin cluster using OpenTofu and provider development
      overrides.
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

      workdir = '/tmp/provider-workflows'
      api_url = 'http://api.vpsadmin.test/'
      location = ${builtins.toJSON location.label}
      os_template = 'debian-latest-x86_64-vpsadminos-minimal'
      public_key = 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIBKuiydHSKEYK5QbvSOgRZpp4mmhUSr9eg9HiTjYpUrY provider-workflows@example.test'

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

      def tofu_state(services, workdir, api_url, token)
        _, output = tofu(services, workdir, api_url, token, 'state pull', timeout: 300)
        JSON.parse(output)
      end

      def state_attrs(state, mode, type, name)
        resource = state.fetch('resources').find do |r|
          r.fetch('mode') == mode && r.fetch('type') == type && r.fetch('name') == name
        end

        expect(resource).not_to be_nil
        resource.fetch('instances').fetch(0).fetch('attributes')
      end

      def managed_state_attrs(state, type, name)
        state_attrs(state, 'managed', type, name)
      end

      def data_state_attrs(state, type, name)
        state_attrs(state, 'data', type, name)
      end

      def tofu_state_rm(services, workdir, api_url, token, *addresses)
        tofu(
          services,
          workdir,
          api_url,
          token,
          "state rm #{Shellwords.join(addresses)}",
          timeout: 300
        )
      end

      def tofu_import(services, workdir, api_url, token, address, id)
        tofu(
          services,
          workdir,
          api_url,
          token,
          "import -input=false #{Shellwords.escape(address)} #{Shellwords.escape(id.to_s)}",
          timeout: 1800
        )
      end

      def expect_no_diff(services, workdir, api_url, token)
        tofu(
          services,
          workdir,
          api_url,
          token,
          'plan -detailed-exitcode -input=false -no-color',
          timeout: 600
        )
      end

      def expect_ip_value(value)
        expect(value).to be_a(String)
        expect(value).not_to eq("")
        value
      end

      def expect_absent_or_deleted(snapshot, name)
        row = snapshot.fetch(name)
        expect(row.nil? || row.fetch('object_state') == 'deleted').to eq(true)
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

      def provider_attrs(overrides = {})
        {
          ssh_key_auto_add: false,
          vps_hostname: 'provider-workflows',
          vps_cpu: 1,
          vps_memory: 1024,
          vps_swap: 0,
          vps_diskspace: 4096,
          public_ipv4_count: 1,
          private_ipv4_count: 1,
          public_ipv6_count: 1,
          start_menu_timeout: 3,
          feature_fuse: false,
          feature_kvm: false,
          feature_lxc: false,
          feature_ppp: false,
          feature_tun: false,
          include_ssh_keys: false,
          include_mount: true,
          dataset_refquota: 1024,
          mount_enable: false,
          mount_on_start_fail: 'mount_later',
          nas_dataset_name: 'provider-workflows-nas/provider-workflows-export',
          nas_dataset_quota: 1024,
          export_dataset: true,
          export_enable: true,
          export_root_squash: false,
          export_read_write: true,
          export_sync: true
        }.merge(overrides)
      end

      def write_provider_config(machine, workdir, public_key, location, os_template, attrs)
        ssh_keys_hcl = attrs.fetch(:include_ssh_keys) ? <<~HCL : ""
          ssh_keys = [vpsadmin_ssh_key.provider_it.id]
        HCL

        mount_hcl = attrs.fetch(:include_mount) ? <<~HCL : ""
          resource "vpsadmin_mount" "provider_it" {
            vps           = vpsadmin_vps.provider_it.id
            dataset       = vpsadmin_dataset.provider_it.id
            mountpoint    = "/mnt/provider-workflows-subdataset"
            enable        = #{attrs.fetch(:mount_enable)}
            mode          = "rw"
            on_start_fail = #{attrs.fetch(:mount_on_start_fail).inspect}
          }

          data "vpsadmin_mount" "provider_it" {
            vps      = vpsadmin_vps.provider_it.id
            mount_id = vpsadmin_mount.provider_it.id

            depends_on = [vpsadmin_mount.provider_it]
          }
        HCL

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
            label    = "provider-workflows"
            key      = #{public_key.inspect}
            auto_add = #{attrs.fetch(:ssh_key_auto_add)}
          }

          data "vpsadmin_ssh_key" "provider_it" {
            label = vpsadmin_ssh_key.provider_it.label

            depends_on = [vpsadmin_ssh_key.provider_it]
          }

          resource "vpsadmin_vps" "provider_it" {
            location             = #{location.inspect}
            install_os_template  = #{os_template.inspect}
            hostname             = #{attrs.fetch(:vps_hostname).inspect}
            cpu                  = #{attrs.fetch(:vps_cpu)}
            memory               = #{attrs.fetch(:vps_memory)}
            swap                 = #{attrs.fetch(:vps_swap)}
            diskspace            = #{attrs.fetch(:vps_diskspace)}
            public_ipv4_count    = #{attrs.fetch(:public_ipv4_count)}
            private_ipv4_count   = #{attrs.fetch(:private_ipv4_count)}
            public_ipv6_count    = #{attrs.fetch(:public_ipv6_count)}
            start_menu_timeout   = #{attrs.fetch(:start_menu_timeout)}
            feature_fuse         = #{attrs.fetch(:feature_fuse)}
            feature_kvm          = #{attrs.fetch(:feature_kvm)}
            feature_lxc          = #{attrs.fetch(:feature_lxc)}
            feature_ppp          = #{attrs.fetch(:feature_ppp)}
            feature_tun          = #{attrs.fetch(:feature_tun)}
            #{ssh_keys_hcl}
          }

          data "vpsadmin_vps" "provider_it" {
            vps_id = vpsadmin_vps.provider_it.id

            depends_on = [vpsadmin_vps.provider_it]
          }

          resource "vpsadmin_dataset" "provider_it" {
            name     = "vps''${vpsadmin_vps.provider_it.id}/provider-workflows-subdataset"
            refquota = #{attrs.fetch(:dataset_refquota)}
          }

          data "vpsadmin_dataset" "provider_it" {
            name = vpsadmin_dataset.provider_it.name

            depends_on = [vpsadmin_dataset.provider_it]
          }

          resource "vpsadmin_dataset" "provider_export" {
            name               = #{attrs.fetch(:nas_dataset_name).inspect}
            quota              = #{attrs.fetch(:nas_dataset_quota)}
            export_dataset     = #{attrs.fetch(:export_dataset)}
            export_enable      = #{attrs.fetch(:export_enable)}
            export_root_squash = #{attrs.fetch(:export_root_squash)}
            export_read_write  = #{attrs.fetch(:export_read_write)}
            export_sync        = #{attrs.fetch(:export_sync)}
          }

          data "vpsadmin_dataset" "provider_export" {
            name = vpsadmin_dataset.provider_export.name

            depends_on = [vpsadmin_dataset.provider_export]
          }

          #{mount_hcl}
        HCL
      end

      def create_provider_token(services, admin_user_id)
        services.api_ruby_json(code: <<~RUBY)
          user = User.find(#{admin_user_id})
          user.update!(level: 2)
          user_agent = UserAgent.find_or_create!('provider-workflows integration test')

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
              label: 'provider-workflows'
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
          lifetime.reason = 'provider-workflows integration test cleanup'
          lifetime.add_expiration = 3600
          lifetime.save!

          puts JSON.dump(default_lifetime_value_id: lifetime.id)
        RUBY
      end

      def wait_for_provider_pool_online(services, pool_id)
        wait_until_block_succeeds(name: "pool #{pool_id} online", timeout: 600) do
          _, output = services.vpsadminctl.succeeds(
            args: ['pool', 'show', pool_id.to_s],
            timeout: 60
          )
          output.fetch('pool').fetch('state') == 'online'
        end
      end

      def require_provider_workflow_setup!
        require_provider_workflow_step!(1)
      end

      def require_provider_workflow_step!(step)
        skip "provider workflow step #{step} did not complete" unless @workflow_step.to_i >= step
      end

      def complete_provider_workflow_step!(step)
        @workflow_step = step
      end

      def ensure_vps_address_pools(services, admin_user_id, location)
        services.api_ruby_json(code: <<~RUBY)
          user = User.find(#{admin_user_id})
          location = Location.find_by!(label: #{location.inspect})
          location.update!(has_ipv6: true)
          env = location.environment

          [
            ['ipv4', 'IPv4 address', 0, 64, 1, :object, 'Ip::Free', 4],
            ['ipv4_private', 'Private IPv4 address', 0, 1024, 1, :object, 'Ip::Free', 4],
            ['ipv6', 'IPv6 address', 0, 64, 1, :object, 'Ip::Free', 4]
          ].each do |name, label, min, max, stepsize, resource_type, free_chain, value|
            resource = ClusterResource.find_or_initialize_by(name: name)
            resource.assign_attributes(
              label: label,
              min: min,
              max: max,
              stepsize: stepsize,
              resource_type: resource_type,
              allocate_chain: nil,
              free_chain: free_chain
            )
            resource.save! if resource.changed?

            user_resource = UserClusterResource.find_or_initialize_by(
              user: user,
              environment: env,
              cluster_resource: resource
            )
            user_resource.value = [user_resource.value.to_i, value].max
            user_resource.save! if user_resource.changed? || user_resource.new_record?
          end

          def ensure_network_ip(location, label, address, prefix, split_prefix, ip_version, role, addr)
            network = Network.find_or_initialize_by(address: address, prefix: prefix)
            network.assign_attributes(
              label: label,
              ip_version: ip_version,
              role: role,
              managed: true,
              split_access: :no_access,
              split_prefix: split_prefix,
              purpose: :vps,
              primary_location: location
            )
            network.save! if network.changed?

            loc_net = LocationNetwork.find_or_initialize_by(location: location, network: network)
            loc_net.assign_attributes(
              primary: true,
              priority: 10,
              autopick: true,
              userpick: true
            )
            loc_net.save! if loc_net.changed? || loc_net.new_record?

            ip = IpAddress.find_by(ip_addr: addr)
            if ip.nil?
              ip = IpAddress.register(
                IPAddress.parse(addr + '/' + split_prefix.to_s),
                network: network,
                user: nil,
                location: location,
                prefix: split_prefix,
                size: 1
              )
            end

            {
              network_id: network.id,
              ip_address_id: ip.id,
              addr: ip.ip_addr
            }
          end

          public_ipv4 = ensure_network_ip(
            location,
            'Provider Workflows Public IPv4',
            '198.51.10.0',
            24,
            32,
            4,
            :public_access,
            '198.51.10.10'
          )
          private_ipv4 = ensure_network_ip(
            location,
            'Provider Workflows Private IPv4',
            '198.51.20.0',
            24,
            32,
            4,
            :private_access,
            '198.51.20.10'
          )
          public_ipv6 = ensure_network_ip(
            location,
            'Provider Workflows Public IPv6',
            '2001:db8:100::',
            64,
            128,
            6,
            :public_access,
            '2001:db8:100::10'
          )

          puts JSON.dump(
            public_ipv4: public_ipv4,
            private_ipv4: private_ipv4,
            public_ipv6: public_ipv6
          )
        RUBY
      end

      def workflow_snapshot(services, vps_id:, ssh_key_id:, dataset_id:, nas_dataset_id:, mount_id: nil, export_id: nil)
        mount_id_value = mount_id.nil? ? 'nil' : Integer(mount_id).to_s
        export_id_value = export_id.nil? ? 'nil' : Integer(export_id).to_s

        services.api_ruby_json(code: <<~RUBY)
          def int_or_nil(value)
            value.nil? ? nil : value.to_i
          end

          def host_ip_for(vps, version, role)
            return nil if vps.nil?

            vps.host_ip_addresses
               .joins(ip_address: :network)
               .where(networks: { ip_version: version, role: Network.roles.fetch(role) })
               .where.not(host_ip_addresses: { order: nil })
               .order('host_ip_addresses.`order`')
               .pick(:ip_addr)
          end

          vps = Vps.including_deleted.find_by(id: #{Integer(vps_id)})
          key = UserPublicKey.find_by(id: #{Integer(ssh_key_id)})
          dataset = Dataset.find_by(id: #{Integer(dataset_id)})
          nas_dataset = Dataset.find_by(id: #{Integer(nas_dataset_id)})
          mount_id = #{mount_id_value}
          export_id = #{export_id_value}
          mount = mount_id && Mount.find_by(id: mount_id)
          export = export_id && Export.find_by(id: export_id)

          features = VpsFeature.where(vps_id: #{Integer(vps_id)}).to_h do |feature|
            [feature.name, feature.enabled]
          end

          deployed_key_events = ObjectHistory.where(
            tracked_object_type: 'Vps',
            tracked_object_id: #{Integer(vps_id)},
            event_type: 'deploy_public_key'
          ).count

          puts JSON.dump(
            ssh_key: key && {
              id: key.id,
              label: key.label,
              key: key.key,
              auto_add: key.auto_add,
              fingerprint: key.fingerprint,
              comment: key.comment
            },
            vps: vps && {
              id: vps.id,
              object_state: vps.object_state,
              hostname: vps.hostname,
              cpu: int_or_nil(vps.cpu),
              memory: int_or_nil(vps.memory),
              swap: int_or_nil(vps.swap),
              diskspace: int_or_nil(vps.dataset && vps.dataset.refquota),
              start_menu_timeout: int_or_nil(vps.start_menu_timeout),
              public_ipv4_address: host_ip_for(vps, 4, 'public_access'),
              private_ipv4_address: host_ip_for(vps, 4, 'private_access'),
              public_ipv6_address: host_ip_for(vps, 6, 'public_access'),
              deployed_key_events: deployed_key_events
            },
            features: features,
            dataset: dataset && {
              id: dataset.id,
              full_name: dataset.full_name,
              object_state: dataset.object_state,
              refquota: int_or_nil(dataset.refquota)
            },
            mount: mount && {
              id: mount.id,
              object_state: mount.object_state,
              vps_id: mount.vps_id,
              dataset_id: mount.dataset && mount.dataset.id,
              mountpoint: mount.dst,
              enable: mount.enabled,
              mode: mount.mode,
              on_start_fail: mount.on_start_fail
            },
            nas_dataset: nas_dataset && {
              id: nas_dataset.id,
              full_name: nas_dataset.full_name,
              object_state: nas_dataset.object_state,
              quota: int_or_nil(nas_dataset.quota),
              export_id: nas_dataset.export && nas_dataset.export.id
            },
            export: export && {
              id: export.id,
              object_state: export.object_state,
              enabled: export.enabled,
              root_squash: export.root_squash,
              rw: export.rw,
              sync: export.sync,
              path: export.path,
              ip_address: export.host_ip_address && export.host_ip_address.ip_addr
            }
          )
        RUBY
      end

      def ip_assignment_state(services, *addrs)
        services.api_ruby_json(code: <<~RUBY)
          addrs = #{JSON.dump(addrs)}
          rows = addrs.to_h do |addr|
            ip = IpAddress.find_by(ip_addr: addr)
            host = ip && ip.host_ip_addresses.first
            [
              addr,
              ip && {
                ip_address_id: ip.id,
                network_interface_id: ip.network_interface_id,
                host_order: host && host.order
              }
            ]
          end

          puts JSON.dump(rows)
        RUBY
      end

      def assert_state_for_attrs(state, attrs)
        ssh_key = data_state_attrs(state, 'vpsadmin_ssh_key', 'provider_it')
        vps = data_state_attrs(state, 'vpsadmin_vps', 'provider_it')
        dataset = data_state_attrs(state, 'vpsadmin_dataset', 'provider_it')
        nas_dataset = data_state_attrs(state, 'vpsadmin_dataset', 'provider_export')

        expect(ssh_key.fetch('auto_add')).to eq(attrs.fetch(:ssh_key_auto_add))
        expect(ssh_key.fetch('comment')).to eq('provider-workflows@example.test')
        expect(vps.fetch('hostname')).to eq(attrs.fetch(:vps_hostname))
        expect(vps.fetch('cpu')).to eq(attrs.fetch(:vps_cpu))
        expect(vps.fetch('memory')).to eq(attrs.fetch(:vps_memory))
        expect(vps.fetch('swap')).to eq(attrs.fetch(:vps_swap))
        expect(vps.fetch('diskspace')).to eq(attrs.fetch(:vps_diskspace))
        expect(vps.fetch('start_menu_timeout')).to eq(attrs.fetch(:start_menu_timeout))
        expect(vps.fetch('feature_fuse')).to eq(attrs.fetch(:feature_fuse))
        expect(vps.fetch('feature_kvm')).to eq(attrs.fetch(:feature_kvm))
        expect(vps.fetch('feature_lxc')).to eq(attrs.fetch(:feature_lxc))
        expect(vps.fetch('feature_ppp')).to eq(attrs.fetch(:feature_ppp))
        expect(vps.fetch('feature_tun')).to eq(attrs.fetch(:feature_tun))
        expect(dataset.fetch('refquota')).to eq(attrs.fetch(:dataset_refquota))
        expect(nas_dataset.fetch('quota')).to eq(attrs.fetch(:nas_dataset_quota))
        expect(nas_dataset.fetch('export_dataset')).to eq(attrs.fetch(:export_dataset))
        expect(nas_dataset.fetch('export_enable')).to eq(attrs.fetch(:export_enable)) if attrs.fetch(:export_dataset)
        expect(nas_dataset.fetch('export_root_squash')).to eq(attrs.fetch(:export_root_squash)) if attrs.fetch(:export_dataset)
        expect(nas_dataset.fetch('export_read_write')).to eq(attrs.fetch(:export_read_write)) if attrs.fetch(:export_dataset)
        expect(nas_dataset.fetch('export_sync')).to eq(attrs.fetch(:export_sync)) if attrs.fetch(:export_dataset)

        if attrs.fetch(:include_mount)
          mount = data_state_attrs(state, 'vpsadmin_mount', 'provider_it')
          expect(mount.fetch('enable')).to eq(attrs.fetch(:mount_enable))
          expect(mount.fetch('on_start_fail')).to eq(attrs.fetch(:mount_on_start_fail))
        end
      end

      def assert_snapshot_for_attrs(snapshot, attrs, public_ipv4:, private_ipv4:, public_ipv6:, mount_expected:, export_expected:)
        key = snapshot.fetch('ssh_key')
        expect(key.fetch('auto_add')).to eq(attrs.fetch(:ssh_key_auto_add))
        expect(key.fetch('comment')).to eq('provider-workflows@example.test')

        vps = snapshot.fetch('vps')
        expect(vps.fetch('object_state')).to eq('active')
        expect(vps.fetch('hostname')).to eq(attrs.fetch(:vps_hostname))
        expect(vps.fetch('cpu')).to eq(attrs.fetch(:vps_cpu))
        expect(vps.fetch('memory')).to eq(attrs.fetch(:vps_memory))
        expect(vps.fetch('swap')).to eq(attrs.fetch(:vps_swap))
        expect(vps.fetch('diskspace')).to eq(attrs.fetch(:vps_diskspace))
        expect(vps.fetch('start_menu_timeout')).to eq(attrs.fetch(:start_menu_timeout))
        expect(vps.fetch('public_ipv4_address')).to eq(public_ipv4)
        expect(vps.fetch('private_ipv4_address')).to eq(private_ipv4)
        expect(vps.fetch('public_ipv6_address')).to eq(public_ipv6)

        features = snapshot.fetch('features')
        expect(features.fetch('fuse')).to eq(attrs.fetch(:feature_fuse))
        expect(features.fetch('kvm')).to eq(attrs.fetch(:feature_kvm))
        expect(features.fetch('lxc')).to eq(attrs.fetch(:feature_lxc))
        expect(features.fetch('ppp')).to eq(attrs.fetch(:feature_ppp))
        expect(features.fetch('tun')).to eq(attrs.fetch(:feature_tun))

        dataset = snapshot.fetch('dataset')
        expect(dataset.fetch('object_state')).to eq('active')
        expect(dataset.fetch('refquota')).to eq(attrs.fetch(:dataset_refquota))

        nas_dataset = snapshot.fetch('nas_dataset')
        expect(nas_dataset.fetch('object_state')).to eq('active')
        expect(nas_dataset.fetch('quota')).to eq(attrs.fetch(:nas_dataset_quota))

        if mount_expected
          mount = snapshot.fetch('mount')
          expect(mount.fetch('object_state')).to eq('active')
          expect(mount.fetch('enable')).to eq(attrs.fetch(:mount_enable))
          expect(mount.fetch('mode')).to eq('rw')
          expect(mount.fetch('on_start_fail')).to eq(attrs.fetch(:mount_on_start_fail))
        end

        if export_expected
          export = snapshot.fetch('export')
          expect(export.fetch('object_state')).to eq('active')
          expect(export.fetch('enabled')).to eq(attrs.fetch(:export_enable))
          expect(export.fetch('root_squash')).to eq(attrs.fetch(:export_root_squash))
          expect(export.fetch('rw')).to eq(attrs.fetch(:export_read_write))
          expect(export.fetch('sync')).to eq(attrs.fetch(:export_sync))
          expect(export.fetch('path')).not_to eq("")
          expect(export.fetch('ip_address')).not_to eq("")
        end
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

      describe 'provider workflows', order: :defined do
        it 'prepares vpsAdmin and OpenTofu inputs' do
          hypervisor_pool = create_pool(
            services,
            node_id: node1_id,
            label: 'provider-workflows-hypervisor',
            filesystem: primary_pool_fs,
            role: 'hypervisor'
          )
          wait_for_provider_pool_online(services, hypervisor_pool.fetch('id'))

          primary_pool = create_pool(
            services,
            node_id: node1_id,
            label: 'provider-workflows-primary',
            filesystem: 'tank/provider-workflows-primary',
            role: 'primary',
            properties: { refquota_check: false }
          )
          wait_for_provider_pool_online(services, primary_pool.fetch('id'))

          @nas_root = create_top_level_dataset(
            services,
            admin_user_id: admin_user_id,
            pool_id: primary_pool.fetch('id'),
            dataset_name: 'provider-workflows-nas',
            refquota: nil
          )
          ensure_private_export_network_with_ips(
            services,
            admin_user_id: admin_user_id,
            dataset_id: @nas_root.fetch('dataset_id'),
            count: 1
          )
          ensure_vps_address_pools(services, admin_user_id, location)

          configure_vps_soft_delete_lifetime(services, location)
          @token = create_provider_token(services, admin_user_id).fetch('token')
          @attrs = provider_attrs
          @nas_dataset_name = @attrs.fetch(:nas_dataset_name)
          services.succeeds("rm -rf #{Shellwords.escape(workdir)} && install -d #{Shellwords.escape(workdir)}")
          write_tofu_cli_config(services, workdir)
          write_provider_config(services, workdir, public_key, location, os_template, @attrs)
          complete_provider_workflow_step!(1)
        end

        it 'creates provider resources, reads data sources, and allocates IP addresses' do
          require_provider_workflow_setup!
          tofu(services, workdir, api_url, @token, 'apply -auto-approve -input=false')

          state = tofu_state(services, workdir, api_url, @token)
          @ssh_key_id = managed_state_attrs(state, 'vpsadmin_ssh_key', 'provider_it').fetch('id')
          @vps_id = managed_state_attrs(state, 'vpsadmin_vps', 'provider_it').fetch('id')
          @dataset_id = managed_state_attrs(state, 'vpsadmin_dataset', 'provider_it').fetch('id')
          @mount_id = managed_state_attrs(state, 'vpsadmin_mount', 'provider_it').fetch('id')
          nas_dataset_attrs = managed_state_attrs(state, 'vpsadmin_dataset', 'provider_export')
          @nas_dataset_id = nas_dataset_attrs.fetch('id')
          @export_id = nas_dataset_attrs.fetch('export_id')
          vps_attrs = data_state_attrs(state, 'vpsadmin_vps', 'provider_it')
          @public_ipv4 = expect_ip_value(vps_attrs.fetch('public_ipv4_address'))
          @private_ipv4 = expect_ip_value(vps_attrs.fetch('private_ipv4_address'))
          @public_ipv6 = expect_ip_value(vps_attrs.fetch('public_ipv6_address'))

          expect(@ssh_key_id).to match(/\A[0-9]+\z/)
          expect(@vps_id).to match(/\A[0-9]+\z/)
          expect(@dataset_id).to match(/\A[0-9]+\z/)
          expect(@mount_id).to match(/\A[0-9]+\z/)
          expect(@nas_dataset_id).to match(/\A[0-9]+\z/)
          expect(@export_id).to be_a(Integer)
          assert_state_for_attrs(state, @attrs)

          snapshot = workflow_snapshot(
            services,
            vps_id: @vps_id,
            ssh_key_id: @ssh_key_id,
            dataset_id: @dataset_id,
            mount_id: @mount_id,
            nas_dataset_id: @nas_dataset_id,
            export_id: @export_id
          )
          assert_snapshot_for_attrs(
            snapshot,
            @attrs,
            public_ipv4: @public_ipv4,
            private_ipv4: @private_ipv4,
            public_ipv6: @public_ipv6,
            mount_expected: true,
            export_expected: true
          )
          complete_provider_workflow_step!(2)
        end

        it 'updates resources and converges without further changes' do
          require_provider_workflow_step!(2)
          @attrs = provider_attrs(
            ssh_key_auto_add: true,
            vps_hostname: 'provider-workflows-renamed',
            vps_cpu: 2,
            vps_memory: 1536,
            vps_swap: 0,
            vps_diskspace: 5120,
            start_menu_timeout: 10,
            feature_fuse: true,
            feature_kvm: true,
            feature_lxc: true,
            feature_ppp: true,
            feature_tun: true,
            dataset_refquota: 2048,
            mount_enable: true,
            mount_on_start_fail: 'fail_start',
            nas_dataset_quota: 2048,
            export_enable: false,
            export_root_squash: true,
            export_read_write: false,
            export_sync: false
          )
          write_provider_config(services, workdir, public_key, location, os_template, @attrs)

          tofu(services, workdir, api_url, @token, 'apply -auto-approve -input=false')

          state = tofu_state(services, workdir, api_url, @token)
          assert_state_for_attrs(state, @attrs)
          vps_attrs = data_state_attrs(state, 'vpsadmin_vps', 'provider_it')
          expect(vps_attrs.fetch('public_ipv4_address')).to eq(@public_ipv4)
          expect(vps_attrs.fetch('private_ipv4_address')).to eq(@private_ipv4)
          expect(vps_attrs.fetch('public_ipv6_address')).to eq(@public_ipv6)

          snapshot = workflow_snapshot(
            services,
            vps_id: @vps_id,
            ssh_key_id: @ssh_key_id,
            dataset_id: @dataset_id,
            mount_id: @mount_id,
            nas_dataset_id: @nas_dataset_id,
            export_id: @export_id
          )
          assert_snapshot_for_attrs(
            snapshot,
            @attrs,
            public_ipv4: @public_ipv4,
            private_ipv4: @private_ipv4,
            public_ipv6: @public_ipv6,
            mount_expected: true,
            export_expected: true
          )

          expect_no_diff(services, workdir, api_url, @token)
          complete_provider_workflow_step!(3)
        end

        it 'deletes and recreates a mount without changing the VPS or dataset' do
          require_provider_workflow_step!(3)
          old_mount_id = @mount_id
          without_mount = @attrs.merge(include_mount: false)
          write_provider_config(services, workdir, public_key, location, os_template, without_mount)

          tofu(services, workdir, api_url, @token, 'apply -auto-approve -input=false')
          snapshot = workflow_snapshot(
            services,
            vps_id: @vps_id,
            ssh_key_id: @ssh_key_id,
            dataset_id: @dataset_id,
            mount_id: old_mount_id,
            nas_dataset_id: @nas_dataset_id,
            export_id: @export_id
          )
          expect_absent_or_deleted(snapshot, 'mount')
          expect(snapshot.fetch('vps').fetch('object_state')).to eq('active')
          expect(snapshot.fetch('dataset').fetch('object_state')).to eq('active')
          expect_no_diff(services, workdir, api_url, @token)

          write_provider_config(services, workdir, public_key, location, os_template, @attrs)
          tofu(services, workdir, api_url, @token, 'apply -auto-approve -input=false')
          state = tofu_state(services, workdir, api_url, @token)
          @mount_id = managed_state_attrs(state, 'vpsadmin_mount', 'provider_it').fetch('id')
          expect(@mount_id).to match(/\A[0-9]+\z/)

          snapshot = workflow_snapshot(
            services,
            vps_id: @vps_id,
            ssh_key_id: @ssh_key_id,
            dataset_id: @dataset_id,
            mount_id: @mount_id,
            nas_dataset_id: @nas_dataset_id,
            export_id: @export_id
          )
          assert_snapshot_for_attrs(
            snapshot,
            @attrs,
            public_ipv4: @public_ipv4,
            private_ipv4: @private_ipv4,
            public_ipv6: @public_ipv6,
            mount_expected: true,
            export_expected: true
          )
          expect_no_diff(services, workdir, api_url, @token)
          complete_provider_workflow_step!(4)
        end

        it 'deletes and recreates a dataset export without deleting the dataset' do
          require_provider_workflow_step!(4)
          old_export_id = @export_id
          without_export = @attrs.merge(export_dataset: false)
          write_provider_config(services, workdir, public_key, location, os_template, without_export)

          tofu(services, workdir, api_url, @token, 'apply -auto-approve -input=false')
          state = tofu_state(services, workdir, api_url, @token)
          nas_dataset_attrs = data_state_attrs(state, 'vpsadmin_dataset', 'provider_export')
          expect(nas_dataset_attrs.fetch('export_dataset')).to eq(false)
          expect(nas_dataset_attrs.fetch('export_id')).to eq(0)

          snapshot = workflow_snapshot(
            services,
            vps_id: @vps_id,
            ssh_key_id: @ssh_key_id,
            dataset_id: @dataset_id,
            mount_id: @mount_id,
            nas_dataset_id: @nas_dataset_id,
            export_id: old_export_id
          )
          expect(snapshot.fetch('nas_dataset').fetch('object_state')).to eq('active')
          expect_absent_or_deleted(snapshot, 'export')
          expect_no_diff(services, workdir, api_url, @token)

          write_provider_config(services, workdir, public_key, location, os_template, @attrs)
          tofu(services, workdir, api_url, @token, 'apply -auto-approve -input=false')
          state = tofu_state(services, workdir, api_url, @token)
          @export_id = managed_state_attrs(state, 'vpsadmin_dataset', 'provider_export').fetch('export_id')
          expect(@export_id).to be_a(Integer)

          snapshot = workflow_snapshot(
            services,
            vps_id: @vps_id,
            ssh_key_id: @ssh_key_id,
            dataset_id: @dataset_id,
            mount_id: @mount_id,
            nas_dataset_id: @nas_dataset_id,
            export_id: @export_id
          )
          assert_snapshot_for_attrs(
            snapshot,
            @attrs,
            public_ipv4: @public_ipv4,
            private_ipv4: @private_ipv4,
            public_ipv6: @public_ipv6,
            mount_expected: true,
            export_expected: true
          )
          expect_no_diff(services, workdir, api_url, @token)
          complete_provider_workflow_step!(5)
        end

        it 'imports existing resources and plans without drift' do
          require_provider_workflow_step!(5)
          write_provider_config(services, workdir, public_key, location, os_template, @attrs)
          tofu_state_rm(
            services,
            workdir,
            api_url,
            @token,
            'vpsadmin_mount.provider_it',
            'vpsadmin_dataset.provider_it',
            'vpsadmin_dataset.provider_export',
            'vpsadmin_vps.provider_it',
            'vpsadmin_ssh_key.provider_it'
          )

          tofu_import(services, workdir, api_url, @token, 'vpsadmin_ssh_key.provider_it', @ssh_key_id)
          tofu_import(services, workdir, api_url, @token, 'vpsadmin_vps.provider_it', @vps_id)
          tofu_import(
            services,
            workdir,
            api_url,
            @token,
            'vpsadmin_dataset.provider_it',
            "vps#{@vps_id}/provider-workflows-subdataset"
          )
          tofu_import(
            services,
            workdir,
            api_url,
            @token,
            'vpsadmin_dataset.provider_export',
            @nas_dataset_name
          )
          tofu_import(services, workdir, api_url, @token, 'vpsadmin_mount.provider_it', @mount_id)

          expect_no_diff(services, workdir, api_url, @token)
          complete_provider_workflow_step!(6)
        end

        it 'deploys the configured SSH key to the VPS' do
          require_provider_workflow_step!(6)
          @attrs = @attrs.merge(include_ssh_keys: true)
          write_provider_config(services, workdir, public_key, location, os_template, @attrs)

          tofu(services, workdir, api_url, @token, 'apply -auto-approve -input=false')

          authorized_keys = nil
          wait_until_block_succeeds(name: "public key deployed to VPS #{@vps_id}") do
            authorized_keys = vps_authorized_keys_lines(node, vps_id: @vps_id, timeout: 120)
            authorized_keys.count(public_key) == 1
          end
          expect(authorized_keys.count(public_key)).to eq(1)

          snapshot = workflow_snapshot(
            services,
            vps_id: @vps_id,
            ssh_key_id: @ssh_key_id,
            dataset_id: @dataset_id,
            mount_id: @mount_id,
            nas_dataset_id: @nas_dataset_id,
            export_id: @export_id
          )
          expect(snapshot.fetch('vps').fetch('deployed_key_events')).to be >= 1
          expect_no_diff(services, workdir, api_url, @token)
          complete_provider_workflow_step!(7)
        end

        it 'destroys provider resources and releases assigned IP addresses' do
          require_provider_workflow_step!(7)
          services.vpsadminctl.succeeds(
            args: ['vps', 'stop', @vps_id],
            parameters: { force: true }
          )
          wait_for_vps_on_node(services, vps_id: @vps_id, node_id: node1_id, running: false)

          tofu(services, workdir, api_url, @token, 'destroy -auto-approve -input=false')

          snapshot = workflow_snapshot(
            services,
            vps_id: @vps_id,
            ssh_key_id: @ssh_key_id,
            dataset_id: @dataset_id,
            mount_id: @mount_id,
            nas_dataset_id: @nas_dataset_id,
            export_id: @export_id
          )
          expect(snapshot.fetch('ssh_key')).to be_nil
          expect(snapshot.fetch('vps').fetch('object_state')).to eq('soft_delete')
          expect_absent_or_deleted(snapshot, 'dataset')
          expect_absent_or_deleted(snapshot, 'mount')
          expect_absent_or_deleted(snapshot, 'nas_dataset')
          expect_absent_or_deleted(snapshot, 'export')

          ip_state = ip_assignment_state(services, @public_ipv4, @private_ipv4, @public_ipv6)
          ip_state.each_value do |row|
            expect(row.fetch('network_interface_id')).to be_nil
            expect(row.fetch('host_order')).to be_nil
          end

          @token = nil
        end
      end
    '';
  }
) clusterArgs
