## Create K3s Cluster with Vagrant and VirtualBox

### Prerequisites:

1. **Install Vagrant** – Follow the [Vagrant Install Guide](https://phoenixnap.com/kb/how-to-install-vagrant-on-ubuntu).
2. **Install VirtualBox** – Download from [VirtualBox](https://www.virtualbox.org/).
3. **Vagrant Box** – Use the [bento/debian-12](https://portal.cloud.hashicorp.com/vagrant/discover/bento/debian-12) Vagrant box.

### Steps:

1. **Create Shared Directory**:
   ```bash
   mkdir -p ./Shared
   ```

2. **Start Vagrant Cluster**:
   ```bash
   vagrant up
   ```

3. **Copy K3s Kubeconfig File**:
   ```bash
   cat Shared/k3s.yaml >> ~/.kube/config
   ```

### Notes:

- Access the cluster using `kubectl`:
  ```bash
  kubectl get nodes
  ```
- The `Shared` folder is used to transfer files (e.g., `k3s.yaml`) between VMs.
