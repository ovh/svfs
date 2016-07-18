# Using svfs with hubiC

## Retrieve a token and a storage URL

HubiC doesn't expose a keystone endpoint but provides an API
to directly retrieve a token and a storage URL using your hubiC
credentials.

The [hubiC API](https://api.hubic.com) can show these informations
when calling the `/account/credentials` endpoint.

In order to get these values automatically, hubic allows third-party applications
to use its API. Authentication is achieved using the OAUTH protocol.

SVFS will handle the job of fetching a token from the hubiC API everytime this
is necessary using user-defined applications and their credentials. It comes
with a helper command, `hubic-application` that will handle all the hassle of
registring an application in order to use it with svfs (i.e. setting scope,
getting request token, getting access token and finally getting your refresh token).


## Create an application in your hubiC profile

Go to https://hubic.com/home/browser/developers/ and add an application. Application
name must be unique across hubiC, you can run the `hubic-application` command to have
a unique application name suggested.

## Register this application for SVFS

Note application client ID and client Secret and run the `hubic-application` command.
You will be prompted these informations as well as your email and password, then
minimum required mount options will be shown at the end of the application registration
process.

## Access your hubiC data

Using options given within the previous step, you can for instance mount your default
hubiC container depending on your system.

Using linux :
```
sudo mount -t svfs -o hubic_auth=<hubic_auth>,hubic_token=<hubic_token>,container=default hubic /mountpoint
```

Using OSX :
```
mount_svfs hubic /mountpoint -o hubic_auth=<hubic_auth>,hubic_token=<hubic_token>,container=default
```

You can access another's container data from the HubiC webapp using the following URL :

* `https://hubic.com/home/browser/#containerName/`
