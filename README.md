# DevOps开发 golang 测试

1. 72小时内完成
2. fork本仓库
3. 通过kubebuilder或者手动创建golang的项目；完成自定义CRD MyStatefulSet核心功能的开发（功能同kubernetes StatefulSet），尽可能的完善！
4. 同时还需要实现ValidatingAdmissionWebhook，为MyStatefulSet提供准入验证
5. 要求不能直接引用kubernetes StatefulSet模块源码
6. 单元测试覆盖率必须大于80%，提交PR的时候，请附带上单元测试覆盖率截图或记录
7. 必须包含controller、AdmissionWebhook部署需要helm chart
8. makefile包含完整的编译流程（controller、AdmissionWebhook镜像的编译、helm chart编译、单元测试）
9. 完成以后通过pull request 提交，并备注面试姓名+联系方式，然后即时联系HR以免超时；

谢谢合作

测试说明
========

环境准备
--------
通过kind搭建测试环境，并创建证书目录
```
sudo kind create cluster --image kindest/node:v1.31.0
mkdir -p /tmp/k8s-webhook-server/serving-certs/
```
通过脚本创建证书，并导入
```
➜  devops-golang-test git:(main) ✗ ./scripts/gen-crt.sh -a ecc -d webhook-service.system.svc -n tls
san:DNS:webhook-service.system.svc
algorithm:ecc
------------- gen ca key-----------------------
------------- gen server key-----------------------
Certificate request self-signature ok
subject=C=US, ST=Florida, L=Miami, O=Little Havana, CN=webhook-service.system.svc
Enter Export Password:
Verifying - Enter Export Password:
➜  devops-golang-test git:(main) ✗ ls certs 
ca.crt  ca.key  ca.srl  san.cnf  tls.crt  tls.csr  tls-fullchain.crt  tls.key  tls.p12
➜  devops-golang-test git:(main) ✗ cp certs/tls.crt /tmp/k8s-webhook-server/serving-certs/
➜  devops-golang-test git:(main) ✗ cp certs/tls.key /tmp/k8s-webhook-server/serving-certs/
```
测试环境准备
```
# 安装envtest
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

# 设置准备环境
➜  devops-golang-test git:(main) ✗ setup-envtest use 1.31.0
Version: 1.31.0
OS/Arch: linux/amd64
Path: /home/shawn/.local/share/kubebuilder-envtest/k8s/1.31.0-linux-amd64
➜  devops-golang-test git:(main) ✗ ln -s /home/shawn/.local/share/kubebuilder-envtest/k8s/1.31.0-linux-amd64 bin/k8s/1.31.0-linux-amd64
```

验证
====
生成manifest并本地运行
```
➜  devops-golang-test git:(main) ✗ make manifests 
/home/shawn/Workspaces/devops-golang-test/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
➜  devops-golang-test git:(main) ✗ make install
/home/shawn/Workspaces/devops-golang-test/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/home/shawn/Workspaces/devops-golang-test/bin/kustomize build config/crd | kubectl create -f -
customresourcedefinition.apiextensions.k8s.io/mystatefulsets.devops.github.com created
➜  devops-golang-test git:(main) ✗ make run
/home/shawn/Workspaces/devops-golang-test/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/home/shawn/Workspaces/devops-golang-test/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go run ./cmd/main.go
2024-10-19T16:35:15+08:00       INFO    controller-runtime.builder      Registering a mutating webhook  {"GVK": "devops.github.com/v1, Kind=MyStatefulSet", "path": "/mutate-devops-github-com-v1-mystatefulset"}
2024-10-19T16:35:15+08:00       INFO    controller-runtime.webhook      Registering webhook     {"path": "/mutate-devops-github-com-v1-mystatefulset"}
2024-10-19T16:35:15+08:00       INFO    controller-runtime.builder      Registering a validating webhook        {"GVK": "devops.github.com/v1, Kind=MyStatefulSet", "path": "/validate-devops-github-com-v1-mystatefulset"}
2024-10-19T16:35:15+08:00       INFO    controller-runtime.webhook      Registering webhook     {"path": "/validate-devops-github-com-v1-mystatefulset"}
2024-10-19T16:35:15+08:00       INFO    setup   starting manager
2024-10-19T16:35:15+08:00       INFO    starting server {"name": "health probe", "addr": "[::]:8081"}
2024-10-19T16:35:15+08:00       INFO    controller-runtime.webhook      Starting webhook server
2024-10-19T16:35:15+08:00       INFO    setup   disabling http/2
2024-10-19T16:35:15+08:00       INFO    Starting EventSource    {"controller": "mystatefulset", "controllerGroup": "devops.github.com", "controllerKind": "MyStatefulSet", "source": "kind source: *v1.MyStatefulSet"}
2024-10-19T16:35:15+08:00       INFO    Starting Controller     {"controller": "mystatefulset", "controllerGroup": "devops.github.com", "controllerKind": "MyStatefulSet"}
2024-10-19T16:35:15+08:00       INFO    controller-runtime.certwatcher  Updated current TLS certificate
2024-10-19T16:35:15+08:00       INFO    controller-runtime.webhook      Serving webhook server  {"host": "", "port": 9443}
2024-10-19T16:35:15+08:00       INFO    controller-runtime.certwatcher  Starting certificate watcher
2024-10-19T16:35:15+08:00       INFO    Starting workers        {"controller": "mystatefulset", "controllerGroup": "devops.github.com", "controllerKind": "MyStatefulSet", "worker count": 1}
...
```

测试statefulset创建，修改副本数，删除
-------------------------------------
创建statefulset
```
# 创建MyStatefulSet，副本数3
➜  devops-golang-test git:(main) ✗ kubectl apply -f ./config/samples/devops_v1_mystatefulset.yaml
mystatefulset.devops.github.com/mss created
➜  devops-golang-test git:(main) ✗ kubectl get mystatefulsets.devops.github.com mss -o yaml
apiVersion: devops.github.com/v1
kind: MyStatefulSet
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"devops.github.com/v1","kind":"MyStatefulSet","metadata":{"annotations":{},"labels":{"app.kubernetes.io/managed-by":"kustomize","app.kubernetes.io/name":"devops-golang-test"},"name":"mss","namespace":"default"},"spec":{"gracePeriod":5,"replicas":3,"size":2,"storageClass":"standard","template":{"spec":{"containers":[{"image":"quay.io/opstree/redis:latest","name":"redis","volumeMounts":[{"mountPath":"/data","name":"data"}]}]}}}}
  creationTimestamp: "2024-10-19T08:36:27Z"
  finalizers:
  - github.com/finlizer
  generation: 2
  labels:
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/name: devops-golang-test
  name: mss
  namespace: default
  resourceVersion: "6371"
  uid: 0f010323-c129-4819-8285-4cc78366c1e3
spec:
  gracePeriod: 5
  replicas: 3
  size: 2
  storageClass: standard
  template:
    metadata: {}
    spec:
      containers:
      - image: quay.io/opstree/redis:latest
        name: redis
        resources: {}
        volumeMounts:
        - mountPath: /data
          name: data
```
查看pod，pvc和pv
```
➜  devops-golang-test git:(main) ✗ kubectl get pod
NAME                                   READY   STATUS    RESTARTS   AGE
pod-mss-0                              1/1     Running   0          2m10s
pod-mss-1                              1/1     Running   0          2m10s
pod-mss-2                              1/1     Running   0          2m10s
➜  devops-golang-test git:(main) ✗ kubectl get pvc
NAME        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
pvc-mss-0   Bound    pvc-fe2a051c-b309-4e5a-b6f1-120d119049fe   2          RWO            standard       <unset>                 2m28s
pvc-mss-1   Bound    pvc-28f296b4-42f3-4ec4-b107-68fdfa8c1f60   2          RWO            standard       <unset>                 2m28s
pvc-mss-2   Bound    pvc-161dc8f6-96e1-4075-8581-268c084abb67   2          RWO            standard       <unset>                 2m28s
➜  devops-golang-test git:(main) ✗ kubectl get pv 
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM               STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-161dc8f6-96e1-4075-8581-268c084abb67   2          RWO            Delete           Bound    default/pvc-mss-2   standard       <unset>                          2m27s
pvc-28f296b4-42f3-4ec4-b107-68fdfa8c1f60   2          RWO            Delete           Bound    default/pvc-mss-1   standard       <unset>                          2m28s
pvc-fe2a051c-b309-4e5a-b6f1-120d119049fe   2          RWO            Delete           Bound    default/pvc-mss-0   standard       <unset>                          2m28s
```
修改副本数为4，查看pod，pvc，pv
```
➜  devops-golang-test git:(main) ✗ kubectl edit mystatefulsets.devops.github.com mss
mystatefulset.devops.github.com/mss edited
➜  devops-golang-test git:(main) ✗ kubectl get pod                                  
NAME                                   READY   STATUS    RESTARTS   AGE
pod-mss-0                              1/1     Running   0          4m27s
pod-mss-1                              1/1     Running   0          4m27s
pod-mss-2                              1/1     Running   0          4m27s
pod-mss-3                              1/1     Running   0          12s
➜  devops-golang-test git:(main) ✗ kubectl get pvc
NAME        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
pvc-mss-0   Bound    pvc-fe2a051c-b309-4e5a-b6f1-120d119049fe   2          RWO            standard       <unset>                 4m32s
pvc-mss-1   Bound    pvc-28f296b4-42f3-4ec4-b107-68fdfa8c1f60   2          RWO            standard       <unset>                 4m32s
pvc-mss-2   Bound    pvc-161dc8f6-96e1-4075-8581-268c084abb67   2          RWO            standard       <unset>                 4m32s
pvc-mss-3   Bound    pvc-8249e502-d8d5-44e1-9b24-831da2314ff3   2          RWO            standard       <unset>                 17s
➜  devops-golang-test git:(main) ✗ kubectl get pv 
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM               STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-161dc8f6-96e1-4075-8581-268c084abb67   2          RWO            Delete           Bound    default/pvc-mss-2   standard       <unset>                          4m30s
pvc-28f296b4-42f3-4ec4-b107-68fdfa8c1f60   2          RWO            Delete           Bound    default/pvc-mss-1   standard       <unset>                          4m31s
pvc-8249e502-d8d5-44e1-9b24-831da2314ff3   2          RWO            Delete           Bound    default/pvc-mss-3   standard       <unset>                          16s
pvc-fe2a051c-b309-4e5a-b6f1-120d119049fe   2          RWO            Delete           Bound    default/pvc-mss-0   standard       <unset>                          4m31s
```
修改副本数为2 **缩容时不删除PVC，只有删除statefulset时才清理PVC**
```
➜  devops-golang-test git:(main) ✗ kubectl edit mystatefulsets.devops.github.com mss
mystatefulset.devops.github.com/mss edited
➜  devops-golang-test git:(main) ✗ kubectl get po
NAME                                   READY   STATUS    RESTARTS   AGE
pod-mss-0                              1/1     Running   0          6m
pod-mss-1                              1/1     Running   0          6m
➜  devops-golang-test git:(main) ✗ kubectl get pvc
NAME        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
pvc-mss-0   Bound    pvc-fe2a051c-b309-4e5a-b6f1-120d119049fe   2          RWO            standard       <unset>                 6m4s
pvc-mss-1   Bound    pvc-28f296b4-42f3-4ec4-b107-68fdfa8c1f60   2          RWO            standard       <unset>                 6m4s
pvc-mss-2   Bound    pvc-161dc8f6-96e1-4075-8581-268c084abb67   2          RWO            standard       <unset>                 6m4s
pvc-mss-3   Bound    pvc-8249e502-d8d5-44e1-9b24-831da2314ff3   2          RWO            standard       <unset>                 109s
➜  devops-golang-test git:(main) ✗ kubectl get pv 
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM               STORAGECLASS   VOLUMEATTRIBUTESCLASS   REASON   AGE
pvc-161dc8f6-96e1-4075-8581-268c084abb67   2          RWO            Delete           Bound    default/pvc-mss-2   standard       <unset>                          6m2s
pvc-28f296b4-42f3-4ec4-b107-68fdfa8c1f60   2          RWO            Delete           Bound    default/pvc-mss-1   standard       <unset>                          6m3s
pvc-8249e502-d8d5-44e1-9b24-831da2314ff3   2          RWO            Delete           Bound    default/pvc-mss-3   standard       <unset>                          108s
pvc-fe2a051c-b309-4e5a-b6f1-120d119049fe   2          RWO            Delete           Bound    default/pvc-mss-0   standard       <unset>                          6m3s
```
删除statefulset
```
➜  devops-golang-test git:(main) ✗ kubectl delete mystatefulsets.devops.github.com mss
mystatefulset.devops.github.com "mss" deleted
➜  devops-golang-test git:(main) ✗ kubectl get po
No resources found in default namespace.
➜  devops-golang-test git:(main) ✗ kubectl get pvc
No resources found in default namespace.
➜  devops-golang-test git:(main) ✗ kubectl get pv 
No resources found
```
Helm chart生成
--------------
```
➜  devops-golang-test git:(main) ✗ make chart
/home/shawn/Workspaces/devops-golang-test/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/home/shawn/Workspaces/devops-golang-test/bin/kustomize build config/crd > chart/mystatefulset/templates/resource.yaml
helm package chart/mystatefulset
WARNING: Kubernetes configuration file is group-readable. This is insecure. Location: /home/shawn/.kube/config
WARNING: Kubernetes configuration file is world-readable. This is insecure. Location: /home/shawn/.kube/config
Successfully packaged chart and saved it to: /home/shawn/Workspaces/devops-golang-test/mystatefulset-0.1.0.tgz
➜  devops-golang-test git:(main) ✗ ls -l mystatefulset-0.1.0.tgz 
-rw-r--r-- 1 shawn shawn 60939 Oct 19 16:46 mystatefulset-0.1.0.tgz
```
