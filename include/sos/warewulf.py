from sos.report.plugins import Plugin, IndependentPlugin

class Warewulf(Plugin, IndependentPlugin):

    short_desc = 'Warewulf provisioning server'

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
          "/usr/share/warewulf/overlays/wwinit/rootfs/warewulf/wwclient",
          # cluster-wide SSH host private keys (public keys are kept)
          "/etc/warewulf/keys/*_key"
        ])

        self.add_cmd_output([
            "wwctl node list",
            "wwctl node list -a",
            "wwctl image list -l",
            "wwctl profile list",
            "wwctl profile list -a",
            "wwctl image kernels",
            "wwctl version"
        ])

        self.add_journal(units="warewulfd.service")

    # A key is treated as a secret when its name contains one of these
    # words (e.g. ipmi "password", "password hash", "bind_password",
    # "ldap_passwd", "api_token", "SECRET_KEY"). A bare "pass" is not
    # matched, so keys like "compass" or "bypass_mode" are left alone.
    # "pass" on its own (auth_pass, pass) is matched only when it is not
    # part of a longer word.
    _secret_key = (r"[\w .-]*?(?:password|passwd|passphrase|secret|token"
                   r"|(?<![a-z])pass(?![a-z]))[\w .-]*?")
    # Boolean values (PasswordlessRoot: "true") are kept because they are
    # useful for troubleshooting and are not secrets.
    _not_bool = r"(?![\"']?(?:true|false)[\"']?\s*$)"
    # wwctl field names that hold a secret: the IPMI password, and node
    # or network device tags whose name looks like a secret.
    _secret_field = (r"(?:Ipmi\.Password|(?:NetDevs\[[^\]]*\]\.)?"
                     r"Tags\[%s\])" % _secret_key)

    def postproc(self):
        # YAML files in /etc/warewulf (nodes.conf, auth.conf, backups).
        # Block scalars (key: | or key: >) are left alone.
        self.do_path_regex_sub(
            r"/etc/warewulf/.*",
            r"(?im)^(\s*-?\s*%s\s*:[ \t]*)%s(?![|>]\s*$)\S.*$"
            % (self._secret_key, self._not_bool),
            r"\1********"
        )
        # JSON files in /etc/warewulf (e.g. ha.json "auth_pass")
        self.do_path_regex_sub(
            r"/etc/warewulf/.*",
            r'(?i)("%s"\s*:\s*)"(?:[^"\\]|\\.)*"' % self._secret_key,
            r'\1"********"'
        )
        # Any private keys or certificates elsewhere in /etc/warewulf
        # (e.g. site-specific key directories)
        self.do_file_private_sub(r"/etc/warewulf/.*")

        # wwctl tables: mask everything after the field column(s), so a
        # value containing spaces is fully masked.
        #   node list -a:    NODE    FIELD  SOURCE  VALUE
        #   profile list -a: PROFILE FIELD  VALUE
        for cmd, cols in (("wwctl node list -a", r"\S+\s+"),
                          ("wwctl profile list -a", r"")):
            self.do_cmd_output_sub(
                cmd,
                r"(?im)^(\S+\s+%s\s+%s)%s\S.*$"
                % (self._secret_field, cols, self._not_bool),
                r"\1********"
            )
            # Resources are printed as JSON: "bind_password":"..."
            self.do_cmd_output_sub(
                cmd,
                r'(?i)("%s"\s*:\s*)"(?:[^"\\]|\\.)*"' % self._secret_key,
                r'\1"********"'
            )
