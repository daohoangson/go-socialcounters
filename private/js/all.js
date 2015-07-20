if (typeof jQuery === 'function') {
	(function($){
		var url = '{url}';
		var now = '{now}';
		var css = '{css}';
		var counts = {counts};
		var services = ['Facebook', 'Twitter', 'Google'];

		var $div = $('<div />')
			.addClass('socialcounters');

		for (var i = 0; i < services.length; i++) {
			var service = services[i];
			if (typeof counts[service] === '_undefined') {
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

			var $img = $('<span >')
				.addClass('sc-data-url')
				.text(service)
				.appendTo($service);

			var $count = $('<span />')
				.addClass('sc-count')
				.text(counts[service])
				.appendTo($service)

			$service.appendTo($div);
		}

		var $container = $('.socialcounters-container');
		if ($container.length > 0) {
			$div.appendTo($container)
		} else {
			$div.appendTo($('body'));
		}

		$('<style />')
			.text(css)
			.appendTo($('head'));

	})(jQuery);
}