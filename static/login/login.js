async function handleLoginForm(e) {
    e.preventDefault();

    const formData = new FormData(e.target);
    const password = formData.get("password");
    const email = formData.get("email");
    const data = { email, password };

    try {
        const resp = await fetch("http://localhost:42069/api/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(data)
        });

        if (!resp.ok) {
            throw new Error(`Server error: ${resp.status}`);
        }

        const res = await resp.json();
        console.log("Success:", res);
        alert("Logged in successfully!");
    } catch (err) {
        console.error(err);
        alert("Error happened: " + err.message);
    }
}

document.getElementById("login-form")
    .addEventListener("submit", handleLoginForm);

