provider "azurerm" {
  features {}
}

variable "base_name" {
  type    = string
  default = "nsg-view"
}

resource "random_id" "unique_name" {
  keepers = {
    "base_name" = var.base_name
  }
  byte_length = 8
}

resource "azurerm_resource_group" "rg" {
  name     = var.base_name
  location = "uksouth"
}

resource "azurerm_network_security_group" "nsg" {
  name                = var.base_name
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  security_rule {
    access                     = "Allow"
    destination_address_prefix = "*"
    destination_port_range     = "22"
    direction                  = "Inbound"
    name                       = "ssh"
    priority                   = 100
    protocol                   = "Tcp"
    source_address_prefix      = "*"
    source_port_range          = "*"
  }
}

resource "azurerm_storage_account" "nsglogs" {
  name                = replace(replace(lower(random_id.unique_name.b64_url), "-", ""), "_", "")
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  account_tier              = "Standard"
  account_kind              = "StorageV2"
  account_replication_type  = "LRS"
  enable_https_traffic_only = true
}

# resource "azurerm_network_watcher" "nw" {
#   name                = var.base_name
#   location            = azurerm_resource_group.rg.location
#   resource_group_name = azurerm_resource_group.rg.name
# }

resource "azurerm_network_watcher_flow_log" "logs" {
  name                = var.base_name
  location            = azurerm_resource_group.rg.location
  resource_group_name = "NetworkWatcherRG"

  network_watcher_name      = "NetworkWatcher_uksouth"
  network_security_group_id = azurerm_network_security_group.nsg.id
  storage_account_id        = azurerm_storage_account.nsglogs.id
  enabled                   = true

  retention_policy {
    enabled = true
    days    = 7
  }
}

resource "azurerm_virtual_network" "vnet" {
  name                = var.base_name
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  address_space = ["10.0.0.0/24"]
}

resource "azurerm_subnet" "vm" {
  name                = "vm"
  resource_group_name = azurerm_resource_group.rg.name

  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = ["10.0.0.0/25"]
}

resource "azurerm_subnet_network_security_group_association" "vm" {
  subnet_id                 = azurerm_subnet.vm.id
  network_security_group_id = azurerm_network_security_group.nsg.id
}

resource "azurerm_public_ip" "vm" {
  name                = var.base_name
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  allocation_method = "Dynamic"
}

resource "azurerm_network_interface" "nic1" {
  name                = var.base_name
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  ip_configuration {
    name                          = "ipconfig1"
    subnet_id                     = azurerm_subnet.vm.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.vm.id
  }
}

resource "azurerm_linux_virtual_machine" "vm" {
  name                = var.base_name
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  size                  = "Standard_D2as_v4"
  admin_username        = "tom"
  network_interface_ids = [azurerm_network_interface.nic1.id]

  admin_ssh_key {
    username   = "tom"
    public_key = file("~/.ssh/id_rsa.pub")
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Standard_LRS"
  }

  source_image_reference {
    offer     = "0001-com-ubuntu-server-focal"
    publisher = "Canonical"
    sku       = "20_04-lts-gen2"
    version   = "latest"
  }
}

output public_ip {
  value = azurerm_public_ip.vm.ip_address
}