NodeCnt = 2

Vagrant.configure("2") do |config|
  
#  config.vbguest.installer_options = { allow_kernel_upgrade: true }
  config.vm.box = "ubuntu/focal64"
  config.vm.provider :virtualbox do |vb|
    vb.memory = 4096 
    vb.cpus = 2
  end

  config.vm.provision :shell, privileged: true, inline: $install_common_tools

  config.vm.define "node-master" do |master|
    master.vm.hostname = "node-master"
    master.vm.network "private_network", ip: "192.168.56.30"
    master.vm.network "forwarded_port", guest: 22, host: 60010, id: "ssh"
    master.vm.provision :shell, privileged: true, inline: $provision_master_node
  end

  (1..NodeCnt).each do |i|
    config.vm.define "node-worker#{i}" do |node| 
      node.vm.hostname = "node-worker#{i}"
      node.vm.network "private_network", ip: "192.168.56.#{i + 30}"
      node.vm.network "forwarded_port", guest: 22, host: "#{i + 60010}", id: "ssh"
    end
  end



end

$install_common_tools = <<-SHELL

# ssh password 접속 활성화
sed -i 's/PasswordAuthentication no/PasswordAuthentication yes/g' /etc/ssh/sshd_config
sed -i 's/#PermitRootLogin yes/PermitRootLogin yes/g' /etc/ssh/sshd_config;
systemctl restart sshd.service

# 방화벽 해제
systemctl stop firewalld && systemctl disable firewalld
systemctl stop NetworkManager && systemctl disable NetworkManager

# Swap 비활성화
swapoff -a && sed -i '/ swap / s/^/#/' /etc/fstab


# install docker ######################################################
# https://docs.docker.com/engine/install/ubuntu/
apt-get remove docker docker-engine docker.io containerd runc
apt-get update
apt-get install \
    ca-certificates \
    curl \
    gnupg \
    lsb-release
mkdir -m 0755 -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
apt-get update
apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# using cgroup = systemd
cat > /etc/docker/daemon.json <<EOF
{
  "exec-opts": ["native.cgroupdriver=systemd"],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "100m"
  },
  "storage-driver": "overlay2",
}
EOF

# install container runtime ################################################
# https://kubernetes.io/docs/setup/production-environment/container-runtimes/
cat <<EOF | tee /etc/modules-load.d/k8s.conf
overlay
br_netfilter
EOF

modprobe overlay
modprobe br_netfilter

# sysctl params required by setup, params persist across reboots
cat <<EOF | tee /etc/sysctl.d/k8s.conf
net.bridge.bridge-nf-call-iptables  = 1
net.bridge.bridge-nf-call-ip6tables = 1
net.ipv4.ip_forward                 = 1
EOF
sudo sysctl --system
#####################################################################



# permissive 모드로 SELinux 설정(효과적으로 비활성화)
# setenforce 0
#sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config

# Ubuntu Update
apt-get update

# Hosts 등록
cat << EOF >> /etc/hosts
192.168.56.30 node-master
192.168.56.31 node-worker1
192.168.56.32 node-worker2
EOF

# mkdir /etc/docker
# mkdir -p /etc/systemd/system/docker.service.d

# 도커 재시작
sudo systemctl enable --now docker 
sudo systemctl daemon-reload 
sudo systemctl restart docker


# 쿠버네티스 설치(1.24)
apt-get update
apt-get install -y apt-transport-https ca-certificates curl
curl -fsSLo /usr/share/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg
echo "deb [signed-by=/usr/share/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main" | sudo tee /etc/apt/sources.list.d/kubernetes.list
apt-get update
apt-get install -y kubelet=1.24.0-00 kubeadm=1.24.0-00 kubectl=1.24.0-00
apt-mark hold kubelet kubeadm kubectl


# 이게 꿀팁
rm /etc/containerd/config.toml
systemctl restart containerd


SHELL

$provision_master_node = <<-SHELL


# 쿠버네티스 초기화 명령 실행
# kubeadm init --apiserver-advertise-address 192.168.56.30 --pod-network-cidr=192.168.0.0/16
kubeadm init --apiserver-advertise-address 192.168.56.30 --pod-network-cidr=192.168.0.0/16
kubeadm token create --print-join-command > ~/join.sh

# 환경변수 설정
mkdir -p $HOME/.kube
cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
chown $(id -u):$(id -g) $HOME/.kube/config
export KUBECONFIG=/etc/kubernetes/admin.conf

# Kubectl 자동완성 기능 설치
# ; yum install bash-completion -y
# ; source <(kubectl completion bash)
# ; echo "source <(kubectl completion bash)" >> ~/.bashrc
apt-get install bash-completion -y

# # Calico 설치
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/tigera-operator.yaml
curl https://raw.githubusercontent.com/projectcalico/calico/v3.25.0/manifests/custom-resources.yaml -O
kubectl create -f custom-resources.yaml

# # Dashboard 설치
# kubectl apply -f https://kubetm.github.io/yamls/k8s-install/dashboard-2.3.0.yaml
# nohup kubectl proxy --port=8001 --address=192.168.1.10 --accept-hosts='^*$' >/dev/null 2>&1 &

SHELL