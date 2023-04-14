### cloud-controller-manager
k8s cloud vendor load balancing controller

### What functions are supported
```text
After the user creates a service service through the kubectl command and the type is load balancing, the controller will automatically pull the available ip segment from cmdb and fill it into the service
```
### QuickStart
```text
The project is divided into two modules. For demonstration purposes, cloud provider services are also integrated. The name is: cloud-provider-manager
```
#### 1、Start cloud vendor service
```shell
# Configure the relevant database
cat cmd/cloud-provider-manager/cloud-provider-manager.yml
db:
  user: xxx
  host: xxx
  port: xxx
  password: xxx
  name: cloud_privoder
```
```shell
cd cmd/cloud-provider-manager/ 
go run main.go --config cloud-provider-manager.yml
```

#### 2、Start the load balancing controller
```shell
# Configure the cloud provider interface address
cat cmd/loadbalance-controller/loadbalance.yml
loadbalance:
  bind: "http://localhost:9999/api/v1/cloudprovider/loadbalance/bind"
  released: "http://localhost:9999/api/v1/cloudprovider/loadbalance/unbind"
  list: "http://localhost:9999/api/v1/cloudprovider/loadbalance/list"
region: cdcm21
```
```shell
cd cmd/loadbalance-controller/
go run loadbalance.go --loadbalanceconfig loadbalance.yml --kubeconfig=$HOME/.kube/config
```
