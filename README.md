# DigitalOcean App Test

This is my personal repo for testing code on the DigitalOcean App Platform. It's based on the [DigitalOcean Sample getting started guide](https://github.com/digitalocean/sample-dockerfile). I can't promise any of the code in here works or is safe. 

## Run Locally

To test the application locally:

```sh
docker build -t digitalocean-app-test .
docker run -p 8080:8080 digitalocean-app-test
```

## Deploy to DigitalOcean App Platform

You will need to fork this repository into your own GitHub account; DigitalOcean will want to connect to your GitHub account in order to clone the repo.

 1. Sign in to [cloud.digitalocean.com](https://cloud.digitalocean.com/)
 2. Click **App Platform** under *Manage*
 3. Select **GitHub** in the *Get started* box and click **Create app**
 4. Click **Connect GitHub account**
 5. Follow the on-screen instructions to allow DigitOcean access to your GitHub account. I *highly* recommend only giving it access to selected repositories (e.g. this one).
 6. 

## Delete App from DigitalOcean App Platform

You will want to delete your app lest you want to continue paying the subscription fee.

 1. Sign in to [cloud.digitalocean.com](https://cloud.digitalocean.com/)
 2. Click on your **project** (e.g. "First-Project") in the left navigation bar
 3. Click your **app** (e.g. "digitalocean-app-test")
 3. Click **Actions > Destroy App**

## License

I can't promise any of the code in here works or is safe. But if you really want to use it, it is licensed under the [Zero-Clause BSD license](https://opensource.org/license/0bsd).

> Zero-Clause BSD
> 
> Permission to use, copy, modify, and/or distribute this software for
> any purpose with or without fee is hereby granted.
> 
> THE SOFTWARE IS PROVIDED “AS IS” AND THE AUTHOR DISCLAIMS ALL
> WARRANTIES WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES
> OF MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE
> FOR ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY
> DAMAGES WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN
> AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT
> OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
