# Social Counters
This golang app provides consistent and faster social counters for popular social networks.
Traditionally, website owners have to include scripts from social networks to display their buttons
and share counters (see [Facebook](https://developers.facebook.com/docs/plugins/share-button),
[Twitter](https://about.twitter.com/resources/buttons) and [Google](https://developers.google.com/+/web/share/) documents).
Unfortunately, those buttons do not align well together (different style, inflexible width, etc.) and more importantly,
they often make a lot of extra requests in order to render correctly, generally make it a bad user experience.

This app implement counters fetcher for Facebook, Twitter and Google+ and return everything in one single request 
of about 8KB (gzip) or 15KB (original). With the only script inclusion, it can display all the signature logos of
those networks **and** their counters for the specified url.

## Usage

### all.js (requires jQuery)
After deploying the app to your server, simply include the script in your website to start rendering the counters.

````
<script src="//socialcounters.domain.com/js/all.js?url=http://www.domain.com"></script>
````

#### Parameters

 * `url` (**required**)
 * `ttl` (_optional_): specify the cache max-age value. Useful if you put your server behind a some caching service or CDN.
 This should help you to control the actual load on the app server.
 * `target` (_optional_): by default, `all.js` looks for `.socialcounters-container` in the DOM and insert its buttons
 into that element. You can use `target` param to change the selector which is used to find the container element.
 When no container is found, `all.js` will append into `body`.
