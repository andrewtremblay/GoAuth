$(document).ready(function() {
    ErrorRedirect("/");

    $("#cancel").click(function() {
        window.location.replace("/");
    });
});