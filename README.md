# tweet-destroyer

This Cloud Run container trawls through the authenticated user's twitter timeline and deletes the tweets that match a given set of rules

## Manual setup

**To enable access to the Twitter API**: you need a Twitter token, token secret, and an access token and secret for your profile - these must have read **and write** access. You can generate these via the Twitter [developer portal](https://developer.twitter.com/en). I saved all of these credentials as a single JSON blob in Google's [Secret Manager](https://cloud.google.com/secret-manager).

**To allow the Cloud Run service to read the secret**: you need to give the service account that is running the Cloud Run container _Secret Manager Secret Accessor_ permissions, to read the secret. You can do this in the Google Cloud Secret Manager console. By default, Cloud Run services or jobs [run as the default Compute Engine service account](https://cloud.google.com/run/docs/configuring/service-accounts).

## Create & deploy a Cloud Run service

This service was created with the [kettle-cli](https://github.com/operatorai/kettle-cli), which you can install [using brew](https://github.com/nlathia/kettle-cli#installing-with-brew), and then allows you to start from a [template](https://github.com/nlathia/kettle-templates):

```bash
❯ kettle create golang-gcloud-run
```

The container can then be built and deployed with:

```bash
❯ kettle deploy .
...

🔍  API Endpoint:  https://<long-url-values>.run.app
✅  Deployed!
```

Finally, you can run this manually with:

```bash
❯ curl -X POST -d '{"dry_run": true, "max_iterations": 2}' https://<long-url-values>.run.app
```

Note: by default, this creates a resource that is public and can be accessed by anyone on the Internet. You might want to change this!

