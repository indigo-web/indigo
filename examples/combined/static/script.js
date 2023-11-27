const chars = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890"

function randomChar() {
	return chars.charAt(Math.floor(Math.random() * chars.length))
}

function randomString(n) {
	return Array.from(Array(n)).map(randomChar).join("")
}

function updateBox(e) {
	const box = document.getElementById("textbox")
	box.innerHTML = randomString(3000)
}

// setInterval(updateBox, 100)
document.addEventListener("mousemove", updateBox)
