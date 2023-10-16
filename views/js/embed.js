const defythost = "https://www.youtube-nocookie.com"

var getYTHost = function() {
	if (localStorage.getItem("ythost") === null) {
		return ('; ' + document.cookie).split(`; ythost=`).pop().split(';')[0] || defythost
	} else {
		return localStorage.getItem("ythost")
	}
}

var getInvQuality = function() {
	if (localStorage.getItem("ythost_quality") === null) {
		return ('; ' + document.cookie).split(`; ythost_quality=`).pop().split(';')[0] || "hd720"
	} else {
		return localStorage.getItem("ythost_quality")
	}
}


var setYTHost = function() {
	if (window.localStorage) {
		var ythost = localStorage.getItem("ythost");
		ythost = prompt("Youtube embed domain\n(https://youtube.com or Invidious instance e.g. https://yewtu.be)", getYTHost());
		localStorage.setItem("ythost", ythost || defythost);
	} else {
		var ythost = ('; ' + document.cookie).split(`; ythost=`).pop().split(';')[0] || defythost;
		ythost = prompt("Youtube embed domain\n(https://youtube.com or Invidious instance e.g. https://yewtu.be)", getYTHost());
		document.cookie = "ythost=" + ythost || defythost + "; expires=Fri, 31 Dec 9999 23:59:59 GMT";
	}

	if (window.localStorage) {
		var ythost_quality = localStorage.getItem("ythost_quality");
		ythost_quality = prompt("Invidious embed quality\n(dash, hd720, medium)", ythost_quality || "hd720");
		localStorage.setItem("ythost_quality", ythost_quality || "hd720");
	} else {
		var ythost_quality = ('; ' + document.cookie).split(`; ythost_quality=`).pop().split(';')[0] || "hd720";
		ythost_quality = prompt("Invidious embed quality\n(dash, hd720, medium)", ythost_quality || "hd720");
		document.cookie = "ythost_quality=" + ythost_quality || defythost + "; expires=Fri, 31 Dec 9999 23:59:59 GMT";
	}

}

var ytembed = function(element) {
	id = element.getAttribute('data');
	if (element.textContent == 'Remove') {
		element.parentNode.removeChild(element.nextElementSibling);
		element.textContent = 'Embed';
	} else {
		el = document.createElement('div');
		el.className = 'media-embed';
		el.innerHTML = '<iframe src="' + getYTHost() + '/embed/' +
			id + '?quality=' + getInvQuality() + 
			'" width="854" height="480" frameborder="0" allowfullscreen style="max-width:100%;max-height:100%;"></iframe>';

		element.parentNode.insertBefore(el, element.nextElementSibling);

		element.textContent = 'Remove';
	}
	return false
};

ythost = getYTHost();

document.querySelectorAll('.comment').forEach(function(element) {
	const regExp = /((?:watch\?v=|youtu\.be|youtube.com\/(?:shorts|v|live)\/)([\w-]{11})(?!\")(?:\S*))/g
	element.innerHTML = element.innerHTML.replace(regExp, "$1 <span>[<a href='" + ythost + "/watch?v=$2' data='$2' onclick='return ytembed(this)'>Embed</a>]</span>");
});
