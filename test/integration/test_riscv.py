import os
import subprocess

from vm import QEMU

BASE_IMG = "https://dl.fedoraproject.org/pub/alt/risc-v/release/41/Cloud/riscv64/images/Fedora-Cloud-Base-Generic-41.20250128.riscv64.qcow2"


# XXX: make fixture
def setup_vm(img_path):
    if not os.path.exists(img_path):
        # XXX: make sure this is cached via a GH cache action
        subprocess.check_call(["curl", "-o", img_path, BASE_IMG])
        subprocess.check_call([
            "virt-customize",
            "-a", "riscv64.qcow2",
            "--password", f"fedora:password:{pw}",
        ])
        # virt-customize -a riscv64.qcow2 --firstboot-command 'sed -i "s/^#PermitRootLogin ./PermitRootLogin yes/" /etc/ssh/sshd_config'


def test_integration_riscv64(tmp_path):
    user = "fedora"
    pw = "riscv"

    img_path = "riscv64.qcow2"
    setup_vm(img_path)

    env = os.environ
    env["CGO_ENABLED"] = "0"
    env["GOARCH"] = "riscv64"
    riscv_bin = tmp_path / "build.riscv64"
    subprocess.check_call([
        "go", "build",
        "-o", riscv_bin,
        "-tags", "containers_image_openpgp",
	"../../cmd/build",
    ], env=env)

    with QEMU(img_path, arch="riscv64") as test_vm:
        test_vm.copy_to(user, riscv_bin, "/usr/bin/build", password=pw)
        test_vm.run(user, "dnf install -y osbuild osbuild-depsolve-dnf", password=pw)
        # XXX: ibcli would make this easier
        config_path = tmp_path / "config.json"
        config_path.write_text("{}")
        test_vm.copy_to(user, config_path, "/tmp/config.json", password=pw)
        test_vm.run(user, "mkdir /tmp/repos")
        test_vm.copy_to("../../data/repositories/fedora-41.json", "/tmp/repos", password=pw)
        test_vm.run(user, "build -type container -distro fedora-41 -config", "/tmp/config.json -repositories", "/tmp/repos")
