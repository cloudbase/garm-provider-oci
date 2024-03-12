# Garm External Provider For OCI

The OCI external provider allows [garm](https://github.com/cloudbase/garm) to create Linux and Windows runners on top of OCI virtual machines.

## Build

Clone the repo:

```bash
git clone https://github.com/cloudbase/garm-provider-oci
```

Build the binary:

```bash
cd garm-provider-oci
go build .
```

Copy the binary on the same system where garm is running, and [point to it in the config](https://github.com/cloudbase/garm/blob/main/doc/providers.md#the-external-provider).

## Configure

The config file for this external provider is a simple toml used to configure the OCI credentials it needs to spin up virtual machines.

```bash
availability_domain = "mQqX:US-ASHBURN-AD-2"
compartment_id = "ocid1.compartment.oc1...fsbq"
subnet_id = "ocid1.subnet.oc1.iad....feoplaka"
network_security_group_id = "ocid1.networksecuritygroup....pfzya"
tenancy_id = "ocid1.tenancy.oc1..aaaaaaaajds7tbqbvrcaiavm2uk34t7wke7jg75aemsacljymbjxcio227oq"
user_id = "ocid1.user.oc1...ug6l37u6a"
region = "us-ashburn-1"
fingerprint = "38...6f:bb"
private_key_path = "/home/ubuntu/.oci/private_key.pem"
private_key_password = ""
```

## Creating a pool

After you [add it to garm as an external provider](https://github.com/cloudbase/garm/blob/main/doc/providers.md#the-external-provider), you need to create a pool that uses it. Assuming you named your external provider as ```oci``` in the garm config, the following command should create a new pool:

```bash
garm-cli pool create \
    --os-type windows \
    --os-arch amd64 \
    --enabled=true \
    --flavor VM.Standard.E4.Flex \
    --image ocid1.image.oc1.iad.aaaaaaaamf7b6c6kvz2itjyflse6ibax2dgmqts2jlahl2zl3mbxlakv4h5a \
    --min-idle-runners 1 \
    --repo 26ae13a1-13e9-47ec-92c9-1526084684cf \
    --tags oci,windows \
    --provider-name oci
```

This will create a new Windows runner pool for the repo with ID `26ae13a1-13e9-47ec-92c9-1526084684cf` on OCI, using the image specified by its OCID `ocid1.image.oc1.iad.aaaaaaaamf7b6c6kvz2itjyflse6ibax2dgmqts2jlahl2zl3mbxlakv4h5a` corresponding to **Windows-Server-2022-Standard-Edition-VM-2024.01.09-0** for the region **US-ASHBURN-1**.

Here is an example for a Linux pool that uses the image specified by its image name:

```bash
garm-cli pool create \
    --os-type linux \
    --os-arch amd64 \
    --enabled=true \
    --flavor VM.Standard.E4.Flex \
    --image ocid1.image.oc1.iad.aaaaaaaah4rpzimrmnqfaxcm2xe3hdtegn4ukqje66rgouxakhvkaxer24oa \
    --min-idle-runners 0 \
    --repo 26ae13a1-13e9-47ec-92c9-1526084684cf \
    --tags oci,linux \
    --provider-name oci
```

Always find a recent image to use. For example, to see available Windows Server 2022 VM Images, you can access [windows-server-2022-vm](https://docs.oracle.com/en-us/iaas/images/windows-server-2022-vm/).

## Tweaking the provider

Garm supports sending opaque json encoded configs to the IaaS providers it hooks into. This allows the providers to implement some very provider specific functionality that doesn't necessarily translate well to other providers. Features that may exists on OCI, may not exist on Azure or AWS and vice versa.

To this end, this provider supports the following extra specs schema:

```bash
{
    "$schema": "http://cloudbase.it/garm-provider-oci/schemas/extra_specs#",
    "type": "object",
    "properties": {
        "ocpus": {
            "type": "number",
            "description": "Number of OCPUs"
        },
        "memory_in_gbs": {
            "type": "number",
            "description": "Memory in GBs"
        },
        "boot_volume_size": {
            "type": "number",
            "description": "Boot volume size in GB"
        },
        "ssh_public_keys": {
            "type": "array",
            "description": "List of SSH public keys",
            "items": {
                "type": "string",
                "description": "A SSH public key"
            }
        }
    },
	"additionalProperties": false
}
```

An example of extra specs json would look like this:

```bash
{
    "ocpus": 1,
    "memory_in_gbs": 4,
    "boot_volume_size": 255,
}
```

To set it on an existing pool, simply run:

```bash
garm-cli pool update --extra-specs='{"boot_volume_size" : 100}' <POOL_ID>
```

You can also set a spec when creating a new pool, using the same flag.

Workers in that pool will be created taking into account the specs you set on the pool.
