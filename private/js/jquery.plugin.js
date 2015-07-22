if (typeof jQuery === 'function') {
	jQuery.fn.socialcounters = function(options) {
		options = $.extend(true, {
			url: 'this',
			shorten: false,
			callback: null,
			mapping: {
				Facebook: {
					link: 'facebook-link',
					count: 'facebook-count'
				},
				Twitter: {
					link: 'twitter-link',
					count: 'twitter-count'
				},
				Google: {
					link: 'google-link',
					count: 'google-count'
				}
			},
			links: {
				Facebook: 'http://www.facebook.com/sharer/sharer.php?u=',
				Twitter: 'https://twitter.com/share?url=',
				Google: 'https://plus.google.com/share?url=',
			}
		}, options);

		if (options.url === 'this') {
			options.url = window.location.href;
		}

		if (!options.url) {
			// do not continue without an url
			return;
		}

		var self = this;
		var callback = 'socialcounters_' + options.url.replace(/[^0-9a-z_]/gi, '');
		window[callback] = function(counts) {
			for (var service in counts) {
				if (typeof options.mapping[service] === 'undefined') {
					continue;
				}
				var mapping = options.mapping[service];

				if (!!mapping.link && options.links[service]) {
					var $link = self.find('[rel=' + mapping.link + ']');
					if (!$link.attr('href')) {
						$link.attr('title', service)
							.attr('href', options.links[service] + options.url)
							.attr('target', '_blank')
							.attr('onclick', 'javascript:window.open(this.href,"","menubar=no,toolbar=no,resizable=yes,scrollbars=yes,height=300,width=600");return false;');
					}
				}

				if (!!mapping.count) {
					var $count = self.find('[rel=' + mapping.count + ']');
					var count = counts[service];

					if (options.shorten) {
						if (count >= 1000000) {
							count = (Math.round(count / 1000000.0 * 10) / 10) + 'm';
						} else if (count >= 1000) {
							count = (Math.round(count / 1000.0 * 10) / 10) + 'k';
						}
					} else {
						if (typeof count.toLocaleString === 'function') {
							count = count.toLocaleString();
						}
					}
					$count.text(count);
				}
			}

			if (options.callback) {
				options.callback(options.url, counts);
			}
		};

		$.ajax({
			cache: true,
			dataType: 'jsonp',
			jsonpCallback: callback,
			url: '{jsonUrl}?url=' + options.url
		})

		return this;
	};
}