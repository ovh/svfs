# Using svfs with OVH Public Cloud Storage (PCS)

## Retrieve your credentials

You can either download your credentials from :
- the [OVH manager](https://www.ovh.com/manager)
- the [OVH API](https://api.ovh.com)

### OVH Manager

Log into the OVH Manager using your credentials.
Then select the *cloud* section, click on your cloud project
then go to the *billing* section. From there, select
the *Openstack* tab, and click on the key icon at the right
of the user you want to extract credentials for. Then
choose to download the *openrc* file. Source it and
enter your password when prompted.

### OVH API

Log into the OVH API with your customer account or using your
application credentials. You can retrieve the *openrc* file
by calling `/cloud/project/{serviceName}/user/{userId}/openrc`.
Source this file and enter your password when prompted.

## Mounting containers

Once you have sourced the *openrc* file, your credentials will
be available as environment variables, starting with prefix `OS_`.

You can use them with svfs like this :

Using linux :
```
sudo mount -t svfs -o username=$OS_USERNAME,password=$OS_PASSWORD,tenant=$OS_TENANT_NAME,region=$OS_REGION_NAME pcs /mountpoint
```

Using OSX :
```
mount_svfs pcs /mountpoint -o username=$OS_USERNAME,password=$OS_PASSWORD,tenant=$OS_TENANT_NAME,region=$OS_REGION_NAME
```
