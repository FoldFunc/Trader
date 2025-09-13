async function handleRegisterForm(e) {
    e.preventDefault();

    const formData = new FormData(e.target);
    const password = formData.get("password");
    const confPassword = formData.get("confirm");

    if (password !== confPassword) {
        console.error("Passwords don't match");
        alert("Passwords don't match");
        return;
    }
    const name = formData.get("name");
    const email = formData.get("email");
    const data = { email, password, name };

    try {
        const resp = await fetch("http://localhost:42069/api/register", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(data)
        });

        if (!resp.ok) {
            throw new Error(`Server error: ${resp.status}`);
        }

        const res = await resp.json();
        console.log("Success:", res);
        alert("Registered successfully!");
    } catch (err) {
        console.error(err);
        alert("Error happened: " + err.message);
    }
}

document.getElementById("register-form")
    .addEventListener("submit", handleRegisterForm);

