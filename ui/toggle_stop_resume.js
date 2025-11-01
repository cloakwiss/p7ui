document.addEventListener("DOMContentLoaded", () => {
    const toggleBtn = document.getElementById("toggle-action-btn");

    const states = {
        stop: {
            title: "Stop",
            action: "@post('/stop')",
            icon: `
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="currentColor">
                    <rect width="10" height="10" x="3" y="3" rx="1.5" />
                </svg>
            `
        },
        resume: {
            title: "Resume",
            action: "@post('/resume')",
            icon: `
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" fill="none"
                     stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <polygon points="5 3 19 12 5 21 5 3"></polygon>
                </svg>
            `
        }
    };

    let current = "stop";

    toggleBtn.addEventListener("click", async () => {
        const currentAction = states[current].action;

        if (currentAction.startsWith("@post")) {
            const url = currentAction.match(/'([^']+)'/)[1]; // extract '/stop' or '/resume'
            console.log("Posting to:", url);

            // Send a POST request (adjust if you use a custom handler)
            await fetch(url, { method: "POST" }).catch(err => console.error(err));
        }

        // Toggle the button to the next state
        current = current === "stop" ? "resume" : "stop";
        const state = states[current];

        toggleBtn.setAttribute("title", state.title);
        toggleBtn.innerHTML = state.icon;

        console.log(`Switched to: ${state.title}, next action = ${state.action}`);
    });
});
