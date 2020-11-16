# swan-demo-go
Shared Web Advertising Network (SWAN) - Demo in go of SWAN, SWIFT and OWID

# Deployment

## AWS Elastic Beanstalk

TODO

## Azure App Service

TODO

# Files

Procfile : needed by AWS Elastic Beanstalk to indicate the application executable for web services.
build.bat : builds AWS or Azure packages on Windows ready for manual deployment.
appsettings.json.rename : template application settings ready for Azure and AWS storage or DynameDB keys.
appsettings.dev.json.rename : development app settings template.
.ebextensions/.config.rename : AWS Elastic Beanstalk .config template ready for additional SSL certificates.

Note: .gitignore will ignore appsettings.json and appsettings.dev.json to limit the risk of commits containing access keys.

# Environments

## Visual Studio Code

Use the Command Palette (Ctrl + Shift + P) to running the `Go: Install/Update Tools` to install `gopkgs`, `dlv` and `gopls`.