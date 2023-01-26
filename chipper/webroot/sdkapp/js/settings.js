function showSection(id) {
    var headings = document.getElementsByClassName("toggleable-section");
    for (var i = 0; i < headings.length; i++) {
        headings[i].style.display = "none";
    }
    document.getElementById(id).style.display = "block";
}
