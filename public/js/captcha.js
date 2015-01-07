$(document).ready(function() {

	// Reload 
    $("#reload").click(function() {

        var image = $("#robots").attr("src");

        // When the first time reload is pressed
        if (image.indexOf("time=") == -1) {

            var human = $("#certification").val()
            $("#robots").attr("src", image + "?human=" + human + "&time=" + new Date().getTime());

            return
        }

        var renew = image.substring(0, image.indexOf("time=")) + "time=" + new Date().getTime();
        $("#robots").attr("src", renew)
    });

});

// Will redirect to sign up page, if the time
// of captcha is expired.
function ErrorRedirect(path) {
	$("#robots").error(function() {
    	window.location.replace(path)
	});     
}
