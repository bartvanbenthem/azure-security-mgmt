# Description
Query all the managed Windows and Linux virtual machines within a specific subscription and their update management compliance from the log analytics workspace extension. Export the virtual machines and update management results to comma seperated CSV files.

## prerequisites
Install azure cli: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli?view=azure-cli-latest

## install
``` shell
$ git clone https://github.com/bartvanbenthem/azure-update-mgmt.git
$ sudo cp azure-update-mgmt/tree/master/clitools/bin/linux/* /usr/bin/
```

## Examples

### run on all subscriptions in the tenant with script and use the azcli config for AUTH
``` shell
# ENV variables
export AZURE_MANAGED_BY_TAGGING_KEY='<<managedby-key-name>>'
export AZURE_MANAGED_BY_TAGGING_VALUE='<<managedby-value-name>>'
export AZURE_TENANT_NAME='<<tenant-name>>'
export OUTPUT_FILE_UPDATES='../update-mgmt.csv'
export OUTPUT_FILE_VM='../vm.csv'

# load subscriptions into array
subscriptions=$(az account list --query [].id -o tsv)

# set column names
printf "%s,%s,%s,%s,%s\n" "Name" "workspaceID" "ostype" "UUID" "managedby" > $OUTPUT_FILE_VM
printf "%s,%s,%s,%s,%s,%s\n"  "name" "ostype" "security" "critical" "compliance" "assessed" > $OUTPUT_FILE_UPDATES

# run binary for every subscription
for s in ${subscriptions[@]}; do {
    export AZURE_SUBSCRIPTION_ID="$s"
    azure-update-mgmt
    }; done
```

### run on a specific subscription with the created SPN and ENV var AUTH
set variables for creating app registration
``` shell
$ spname='<<name-spn>>'
$ tenantId=$(az account show --query tenantId -o tsv)
$ subscriptions=('<<subscription-id-01 subscription-id-02 ...>>')
```
    
#### Create the Azure AD application
``` shell
$ applicationId=$(az ad app create \
    --display-name "$spname" \
    --identifier-uris "https://$spname" \
    --query appId -o tsv)
```

#### Update the application group memebership claims
``` shell
$ az ad app update --id $applicationId --set groupMembershipClaims=All
```

#### Create a service principal for the Azure AD application
``` shell
$ az ad sp create --id $applicationId
```

#### Get the service principal secret
``` shell
$ applicationSecret=$(az ad sp credential reset \
    --name $applicationId \
    --credential-description "passwrd" \
    --query password -o tsv)
```

#### Add SPN to the subscriptions as an reader
``` shell
for s in "${subscriptions[@]}"; do {
    az role assignment create --assignee $applicationId --subscription $s --role 'Reader'
}; done
```

#### set environment variables and run
Once the Azure App registration is created set the following environment variables:
``` shell
$ export AZURE_CLIENT_ID='$applicationId'
$ export AZURE_TENANT_ID=$tenantId
$ export AZURE_CLIENT_SECRET='$applicationSecret'
$ export AZURE_SUBSCRIPTION_ID='<<subscription-id>>'
$ export AZURE_MANAGED_BY_TAGGING_KEY='<<managedby-key-name>>'
$ export AZURE_MANAGED_BY_TAGGING_VALUE='<<managedby-value-name>>'
$ export AZURE_TENANT_NAME='<<tenant-name>>'
$ azure-update-mgmt
```
