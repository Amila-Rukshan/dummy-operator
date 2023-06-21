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
