# tweet-destroyer

This Cloud Run service trawls through as much of the authenticated user's twitter timeline that it can and deletes the tweets that match a given set of rules.

## Manual setup

**To enable access to the Twitter API**: you need a Twitter token, token secret, and an access token and secret for your profile - these must have read **and write** access. You can generate these via the Twitter [developer portal](https://developer.twitter.com/en). I saved all of these credentials as a single JSON blob in Google's [Secret Manager](https://cloud.google.com/secret-manager).

**To use Google Cloud services**: you need to enable several APIs, as needed - specifically, for Cloud Run and Cloud Scheduler.

**To allow the Cloud Run service to read the secret**: you need to give the service account that is running the Cloud Run container _Secret Manager Secret Accessor_ permissions, to read the secret. You can do this in the Google Cloud Secret Manager console. By default, Cloud Run services or jobs [run as the default Compute Engine service account](https://cloud.google.com/run/docs/configuring/service-accounts).

## Create & deploy a Cloud Run service

This service was created with the [kettle-cli](https://github.com/operatorai/kettle-cli), which you can install [using brew](https://github.com/nlathia/kettle-cli#installing-with-brew), and then allows you to start from a [template](https://github.com/nlathia/kettle-templates):

```bash
❯ kettle create golang-gcloud-run
```

I modified the template to use Golang 1.19.1.

The service can be built and deployed with:

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

When `dry_run` is set to `true`, not tweets are deleted. Any non-zero value for `max_iterations` limits how many batches of ~200 tweets the service tries to retrieve

Important! By default, this command creates a resource **that is public and can be accessed by anyone on the Internet**. You [will need](https://cloud.google.com/run/docs/triggering/using-scheduler) to change this: when you deploy the service you are using with Cloud Scheduler, make sure you do NOT allow unauthenticated invocations. 


## How much does this cost?

Cloud Scheduler gives you [3 free jobs per month, per billing account](https://cloud.google.com/scheduler/pricing). A job is not billed for individual executions.

Cloud Run [is free](https://cloud.google.com/run/pricing) up to 180,000 vCPU seconds/month, 360,000 GiB-seconds per month, and 2 million requests.

Cloud Logging [is free](https://cloud.google.com/stackdriver/pricing) up to the first 50 GiB/project ingested per month. Logs retained for the default period don't incur a storage cost.

