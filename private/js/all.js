if (typeof jQuery === 'function') {
	(function($){
		var url = {url};
		var target = {target};
		var shorten = {shorten};
		var now = '{now}';
		var css = '{css}';
		var svgs = {
			'Facebook': {facebooksvg},			
			'Twitter': {twittersvg},
			'Google': {googlesvg},
		};
		var counts = {counts};
		var services = ['Facebook', 'Twitter', 'Google'];

		var supportsSvg = function() {
			// http://stackoverflow.com/questions/654112/how-do-you-detect-support-for-vml-or-svg-in-a-browser
			return document.implementation.hasFeature("http://www.w3.org/TR/SVG11/feature#Shape", "1.0")
		};

		var formatCount = function(count) {
			var unit = '';

			if (shorten) {
				if (count >= 1000000) {
					count = (Math.round(count / 1000000.0 * 10) / 10);
					unit = 'm';
				} else if (count >= 1000) {
					count = (Math.round(count / 1000.0 * 10) / 10);
					unit = 'k';
				}
			}
			if (typeof count.toLocaleString === 'function') {
				count = count.toLocaleString();
			}

			return count + unit;
		}

		var $div = $('<div />')
			.addClass('socialcounters');

		for (var i = 0; i < services.length; i++) {
			var service = services[i];
			if (typeof counts[service] === 'undefined') {
				continue;
			}

			var href = '';
			switch (service) {
				case 'Facebook':
					href = 'http://www.facebook.com/sharer/sharer.php?u=' + url;
					break;
				case 'Twitter':
					href = 'https://twitter.com/share?url=' + url;
					break;
				case 'Google':
					href = 'https://plus.google.com/share?url=' + url;
					break;
			}
			if (!href) {
				continue;
			}

			var $service = $('<a />')
				.addClass('sc-service')
				.addClass('sc-' + service.toLowerCase())
				.attr('title', service)
				.attr('href', href)
				.attr('target', '_blank')
				.attr('onclick', 'javascript:window.open(this.href,"","menubar=no,toolbar=no,resizable=yes,scrollbars=yes,height=300,width=600");return false;');

			var $img;
			if (supportsSvg()
				&& typeof svgs[service] !== 'undefined'
				&& !!svgs[service]
			) {
				$img = $(svgs[service])
					.attr('class', 'sc-logo sc-svg');
			} else {
				$img = $('<span >')
					.addClass('sc-logo')
					.text(service);
			}

			var $count = $('<span />')
				.addClass('sc-count')
				.text(formatCount(counts[service]));

			$img.appendTo($service);
			$count.appendTo($service);
			$service.appendTo($div);
		}

		var $container = $(target);
		if ($container.length > 0) {
			$div.appendTo($container)
		} else {
			$div.appendTo($('body'));
		}

		var $head = $('head');
		if ($head.children('style.sc-css').length == 0) {
			$('<style />')
				.addClass('sc-css')
				.text(css)
				.appendTo($head);
		}

	})(jQuery);
}