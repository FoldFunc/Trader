// getName.js

async function fetchName() {
    try {
        const resp = await fetch("http://localhost:42069/api/fetch/name", {
            method: "GET",
            headers: { "Content-Type": "application/json" },
        });
        if (!resp.ok) throw new Error(`Error: ${resp.status}`);
        const res = await resp.json();
        console.log("Name: ", res.message);
        return res.message;
    } catch (err) {
        console.error(err);
        throw new Error(`Error happened: ${err}`);
    }
}

async function getName() {
    const name = await fetchName();
    document.getElementById("greeting").innerText = `Hello ${name}`;
}

// Run after DOM is ready
document.addEventListener("DOMContentLoaded", async () => {
    await getName();
});

