function updateColor(id) {
  console.log(id);
  var body_styles = window.getComputedStyle(
    document.getElementsByTagName("body")[0]
  );
  var fgColor = body_styles.getPropertyValue("--fg-color");
  var bgColorAlt = body_styles.getPropertyValue("--gg-color-alt");

  l_id = id.replace("section", "icon");
  let elements = document.getElementsByName("icon");
  for (var i = 0; i < elements.length; i++) {
    elements[i].style.color = bgColorAlt;
  }
  var icon = document.getElementById(l_id);
  if (icon) {
    icon.style.color = fgColor;
  }
}
