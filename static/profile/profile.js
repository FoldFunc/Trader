async function checkLogged() {
    try {
        const resp = await fetch("http://localhost:42069/api/logcheck", {
            method: "GET",
            headers: { "Content-Type": "application/json" },
        });
        if (!resp.ok) {
            throw new Error(`Server error ${resp.status}`);
        }
        const res = await resp.json();
        if (res.message == "false") {
            console.log("Not logged in");
            return false;
        }
        console.log("Logged in");
        return true;
    } catch (err) {
        console.error(err)
        throw new Error(`Error happend: ${err}`)
    }
}
document.addEventListener("DOMContentLoaded", async () => {
    const logged = await checkLogged();
    if (logged) {
        document.getElementById("loggedInContent").style.display = "block";
    } else {
        document.getElementById("notLoggedInContent").style.display = "block";
    }
});

