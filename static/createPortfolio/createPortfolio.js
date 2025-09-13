async function handleCreatePortfolioForm(e) {
    e.preventDefault();
    const formData = new FormData(e.target);
    const portfolioName = formData.get("name")
    const data = { portfolioName }
    try {
        const resp = await fetch("http://localhost:42069/api/createPortfolio", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(data)
        });
        if (!resp.ok) {
            throw new Error(`Error happend: ${resp.status}`)
        }
        const res = await resp.json();
        console.log("Portfolio created: ", res)
    } catch (err) {
        console.error(`Error happend: ${err}`)
        throw new Error(`Error happend: ${err}`)
    }
}
document.getElementById("createPortfolio-form")
    .addEventListener("submit", handleCreatePortfolioForm);
