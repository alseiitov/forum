var uploadField = document.getElementById("image_input");

uploadField.onchange = function () {
  if (this.files[0].size > 20 * 1024 * 1024) {
    alert("File is too big!");
    this.value = "";
  }

  if (/\.(jpg|jpeg|jpe|jif|jfif|jfi|png|gif)$/i.test(this.files[0].name.toLowerCase()) === false) {
    alert("Not an image!");
    this.value = "";
  }
};
