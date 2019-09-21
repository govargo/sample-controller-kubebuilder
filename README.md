# Sample Controller(Foo Controller)

This Controller is developed by Kubebuilder v2.0.0

技術書典7で頒布した「実践入門 Kubernetesカスタムコントローラへの道」の第五章に掲載したサンプルコード用のリポジトリです。  
このブランチは、第五章に掲載したサンプルコードとは異なる実装を参考として掲載しています。  

masterブランチでは、`controllerutil.CreateOrUpdate`を使った実装でReconcileし、  
feature/sample-controllerブランチでは、`controllerutil.CreateOrUpdate`を使わない実装でReconcileを行います。

## What is this Controller

This Controller reconcile foo custom resource.

Foo resource owns deployment object.

We define foo.spec.deploymentName & foo.spec.replicas.
When we apply foo, then deployment owned by foo is created.

If we delete deployment object like 'kubectl delete deployment/\<deployment name\>',
foo reconcile it and the deployment is created again.

Or if we scale deployment object like 'kubectl scale deployment/\<deployment name\> --replicas 0',
foo reconcile it and the deployment keeps the foo.spec.replicas.

## How to run

Clone source code

```
$ mkdir -p $GOPATH/src/github.com/govargo
$ cd $GOPATH/src/github.com/govargo
$ git clone https://github.com/govargo/sample-controller-kubebuilder.git
$ cd sample-controller-kubebuilder
```

Run localy

```
$ make install
$ make run
$ kubectl apply -f config/samples/samplecontroller_v1alpha1_foo.yaml
```

Run container as Deployment

```
$ make install
$ make deploy
$ kubectl apply -f config/samples/samplecontroller_v1alpha1_foo.yaml
```

## Reference

This project is inspired by

 * https://github.com/kubernetes/sample-controller
 * https://github.com/jetstack/kubebuilder-sample-controller
