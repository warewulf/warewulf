# Configuration file for the Sphinx documentation builder.
#
# For the full list of built-in configuration values, see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html

# -- Project information -----------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#project-information

project = 'Warewulf User Guide'
copyright = '2024, Warewulf Project Contributors'
author = 'Warewulf Project Contributors'
release = 'main'

# -- General configuration ---------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#general-configuration

extensions = [
    'sphinx.ext.graphviz',
    'sphinx_reredirects',
]

redirects = {
    'contents/background': '../getting-started/introduction.html',
    'contents/glossary': '../getting-started/glossary.html',
    'contents/introduction': '../getting-started/introduction.html',
    'contents/provisioning': '../getting-started/provisioning.html',
    'contents/stateless': '../getting-started/provisioning.html',
    'quickstart/debian12': '../getting-started/debian-quickstart.html',
	'quickstart/el': '../getting-started/el-quickstart.html',
    'quickstart/suse15': '../getting-started/suse-quickstart.html',
    'contents/images': 'images/images.html',
    'contents/kernel': 'images/kernel.html',
    'contents/disks': 'nodes/disks.html',
	'contents/ipmi': 'nodes/ipmi.html',
	'contents/nodeconfig': 'nodes/nodes.html',
	'contents/overlays': 'overlays/overlays.html',
	'contents/templating': 'overlays/templates.html',
	'contents/profiles': 'nodes/profiles.html',
	'contents/boot-management': 'server/bootloaders.html',
	'contents/configuration': 'server/configuration.html',
	'contents/dnsmasq': 'server/dnsmasq.html',
	'contents/initialization': 'server/configuration.html',
	'contents/installation': 'server/installation.html',
	'contents/security': 'server/security.html',
	'contents/setup': 'getting-started/network.html',
	'contents/upgrade': 'server/upgrade.html',
	'contents/wwctl': 'server/wwctl.html',
	'contents/known-issues': 'troubleshooting/known-issues.html',
	'contents/troubleshooting': 'troubleshooting/troubleshooting.html',
    'contributing/development-environment-devcontainer': 'contributing/development-environment',
    'contributing/development-environment-kvm': 'contributing/development-environment',
    'contributing/development-environment-vagrant': 'contributing/development-environment',
    'contributing/development-environment-vbox': 'contributing/development-environment',
}

templates_path = ['_templates']
exclude_patterns = ['_build', 'Thumbs.db', '.DS_Store']



# -- Options for HTML output -------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#options-for-html-output

html_theme = 'sphinx_rtd_theme'
html_static_path = ['_static']

html_theme_options = {
    'sticky_navigation': True,
    'includehidden': True,
    'navigation_depth': 5,
    'prev_next_buttons_location': 'bottom',
    'style_external_links': True,
}

html_context = {
    'display_github': True,
    'github_user': 'warewulf',
    'github_repo': 'warewulf',
    'github_version': 'main',
    'conf_py_path': '/userdocs/',
}

html_logo = 'logo.png'
html_favicon = 'favicon.png'
html_show_copyright = False
