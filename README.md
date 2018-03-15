# terraform-provider-ibm-api
Added the helm chart for the Micro service that exposes REST API for IBM Cloud terraform provider plugin

* List of parameter which can be provided by user while installing the helm chart

|Variable Name|Description|Default Value|
|-------------|-----------|-------------|
|image.repository|docker image for api service||
|image.pullPolicy|Image Pull Policy|IfNotPresent|
|pvc.enabled| Set to true to create pvc (used to store the configuration repo and logs and state file.)|true|
|pvc.storageClass|pvc storage class|ibmc-file-bronze|
|pvc.mountPath|path where the configuration repo will get cloned and used to store state file and logs|/tmp|
|pvc.name|Name of pvc storage|tfstate-data|
|ingress.path|Ingress Path|/|
|ingress.host|ingress hostname value can be obtained from IBM Cloud cluster domain||

## How to use this helm chart

* helm repo add terraformibmprovider-charts  https://ibmterraform.github.io/terraform-provider-ibm-api/

* To get the availale charts in the repo 
    helm search

* helm install <chartname>  --set param=value

for eg: helm install <chartname>  --set pvc.enabled=false --set ingress.host=<ingress_hostname>

 
