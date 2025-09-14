async function getPortfolioInfo() {
    try {
        const resp = await fetch("http://localhost:42069/api/fetch/portfolios", {
            method: "GET",
            headers: { "Content-Type": "application/json" },
        });
        if (!resp.ok) {
            throw new Error(`Error happend: ${resp.status}`);
        }
        const res = await resp.json();
        return res.message.portfolios; // array of portfolios
    } catch (err) {
        console.error("Error happend: ", err);
        return [];
    }
}

async function populateCards() {
    const portfolios = await getPortfolioInfo();
    const container = document.getElementById("portfolio-container");

    container.innerHTML = ""; // clear old content

    portfolios.forEach(p => {
        const card = document.createElement("div");
        card.classList.add("portfolio-card");

        card.innerHTML = `
            <h3>${p.name}</h3>
            <p><strong>Money:</strong> $${p.money}</p>
            <p><strong>Stocks:</strong> ${p.stocks || "None"}</p>
        `;

        container.appendChild(card);
    });
}

populateCards();

