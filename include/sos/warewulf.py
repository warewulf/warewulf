from sos.report.plugins import Plugin, IndependentPlugin

class Warewulf(Plugin, IndependentPlugin):

    short_desc = 'Warewulf provisioning platform'
    
    plugin_name = 'warewulf'
    services = ('warewulfd')
    packages = ('warewulf')

    def setup(self):
       
        self.add_copy_spec([
          "/var/lib/warewulf/overlays/",
          "/usr/share/warewulf/overlays/",
          "/var/log/warewulfd.log",
          "/etc/warewulf/",
          "/var/lib/dhcpd/",
          "/etc/dhcp/",
          "/etc/dnsmasq.d",
          "/var/lib/dnsmasq/dnsmasq.leases"
        ])
        
        self.add_forbidden_path([
          "/var/lib/warewulf/overlays/wwinit/rootfs/warewulf/wwclient",
          "/usr/share/warewulf/overlays/wwinit/rootfs/warewulf/wwclient"
        ])

        self.add_cmd_output([
            "wwctl node list",
            "wwctl node list -a",
            "wwctl container list",
            "wwctl profile list",
            "wwctl profile list -a",
            "wwctl image kernels",
            "wwctl version"
        ])
        
        self.add_journal(units="warewulfd.service")
