# Configuration file for the Sphinx documentation builder.
#
# For the full list of built-in configuration values, see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html

# -- Project information -----------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#project-information

project = 'Warewulf User Guide'
copyright = '2023, Warewulf Project Contributors'
author = 'Warewulf Project Contributors'
release = 'development'

# -- General configuration ---------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#general-configuration

extensions = ['sphinx.ext.autosectionlabel','sphinx.ext.graphviz']

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
    'github_user': 'hpcng',
    'github_repo': 'warewulf',
    'github_version': 'development',
    'conf_py_path': '/userdocs/',
}

html_logo = 'logo.png'
html_favicon = 'favicon.png'
html_show_copyright = False
