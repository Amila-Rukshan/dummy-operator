Execute the following commands from the root directory of the project to test the dummy controller:

1. Set the context in the `kubectl` config to the k8s cluster to run the controller:
    ```
    kubectl config set-context <your-cluster-context>
    ```

2. Install the controller artifacts:
    ```
    make deploy
    ```

3. Apply the sample dummy resource:
    ```
    kubectl apply -f config/samples/dummy.yaml 
    ```

4. Remove the sample resource to test pod cleaning up:
    ```
    kubectl delete -f config/samples/dummy.yaml 
    ```
### Other test scenarios

1. When a pod already exists with the same name as the `Dummy` resource.

    * Create a pod with the same name as the Dummy resource and which is not running nginx pod:
        ```
        kubectl run dumm1 --image httpd
        ```
    * Then, apply the sample dummy resource:
        ```
        kubectl apply -f config/samples/dummy.yaml 
        ```
    Pod will be updated to run nginx container.

    * Deletion of `Dummy` resource will remove the controlled pod as well.

2. After applying the `Dummy` resource, delete the pod it has created. A new pod will be created. 

3. Change the image of the pod created by the `Dummy` resource, it will be reverted back to run nginx:latest container.
