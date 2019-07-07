resource "null_resource" "bastion_test" {
  # Ensure we can SSH as root for the goss tests and also for preturb.sh
  connection {
    host    = "${var.bastion_public_ip}"
    type    = "ssh"
    user    = "root"
    # Increase the timeout so the server has time to reboot
    timeout = "10m"
  }

  provisioner "file" {
    source      = "scripts/run-goss.sh"
    destination = "/root/run-goss.sh"
  }

  provisioner "file" {
    source      = "test/bastion.yaml"
    destination = "/root/goss.yaml"
  }

  provisioner "remote-exec" {
    inline = [
      "chmod +x /root/run-goss.sh",
      "/root/run-goss.sh",
    ]
  }
}