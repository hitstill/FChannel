var getYTHost = function() {
    if (localStorage.getItem("ythost") === null) {
        return ('; ' + document.cookie).split(`; ythost=`).pop().split(';')[0] || "https://yewtu.be"
    } else {
        return localStorage.getItem("ythost")
    }
}

var setYTHost = function() {
    if (window.localStorage) {
        var ythost = localStorage.getItem("ythost");
        ythost = prompt("Youtube embed domain\n(https://youtube.com or Invidious instance e.g. https://yewtu.be)", getYTHost());
        localStorage.setItem("ythost", ythost || "https://yewtu.be");
    } else {
        var ythost = ('; ' + document.cookie).split(`; ythost=`).pop().split(';')[0] || "https://yewtu.be";
        ythost = prompt("Youtube embed domain\n(https://youtube.com or Invidious instance e.g. https://yewtu.be)", getYTHost());
        document.cookie = "ythost=" + ythost || "https://yewtu.be" + "; expires=Fri, 31 Dec 9999 23:59:59 GMT";
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
            id +
            '" width="640" height="360" frameborder="0" allowfullscreen></iframe>';

        element.parentNode.insertBefore(el, element.nextElementSibling);

        element.textContent = 'Remove';
    }
    return false
};

ythost = getYTHost();

document.querySelectorAll('.comment').forEach(function(element) {
    //const regExp = /((?:\/(?:[^\/]+\/.+\/|(?:v|e(?:mbed)?)\/|.*[?&]v=)|youtu\.be\/)([^"&?\/\s]{11}))/gi;
    //const regExp = /watch\?v=(.*?((?=[&#?])|$))/gm;
    const regExp = /((?:watch\?v=|youtu\.be)([\w-]{11})(?!\")(?:\S*))/g
    element.innerHTML = element.innerHTML.replace(regExp, '$1 <span>[<a href="' + ythost + '/watch?v=$2" data="$2" onclick="return ytembed(this)">Embed</a>]</span>');
});
