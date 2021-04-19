function hello() {
    alert("hello");
}


document.addEventListener('DOMContentLoaded', function() {
    document.getElementById('url').value = window.location.href;
    document.getElementById('hello').addEventListener('click', hello);
});
