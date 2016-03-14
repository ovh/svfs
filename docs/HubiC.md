# Usage on hubiC

## Retrieve a token and a storage URL

HubiC doesn't expose a keystone endpoint but provides an API
to directly retrieve a token and a storage URL using your hubiC
credentials.

To retrieve these informations, log into the [hubiC API](https://api.hubic.com)
using your hubiC credentials. Then send a request to `/account/credentials` in
order to get `token` and `endpoint` values.

## Mounting containers

You can then mount your hubiC default container like this :

```
sudo mount -t svfs -o token=<token>,storage_url=<endpoint>,container=default hubic /mountpoint
```

Note : your token will expire after 24 hours, after this time you will need a new one. We will
limitation in the future.

For the moment, you can call the hubiC API from an [application](https://hubic.com/home/browser/apps/)
registered in your hubiC account to automatically remount your storage space with a fresh token.
