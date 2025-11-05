
            document.addEventListener("click", function(event) {
            const dropdown = document.getElementById("target-dropdown");

            if (dropdown.hasAttribute("open") && !dropdown.contains(event.target)) {
                dropdown.removeAttribute("open");
            }
            });

            document.addEventListener("DOMContentLoaded", () => {
            const consoleBody = document.getElementById("console");
            const hookBody = document.getElementById("hooks");


                const observer = new MutationObserver(() => {
                    const console_paneContent = consoleBody.closest(".pane-content");
                    const hooks_paneContent = hookBody.closest(".pane-content");

                    if (console_paneContent) {
                        console_paneContent.scrollTop = console_paneContent.scrollHeight;
                    }
                    if (hooks_paneContent) {
                        hooks_paneContent.scrollTop = hooks_paneContent.scrollHeight;
                    }
                });

                observer.observe(consoleBody, { childList: true });
                observer.observe(hookBody, { childList: true });
            });

    let slashPressed = false;

    document.addEventListener('keydown', function(event) {
        if (event.key === '/') {
            slashPressed = true;
            event.preventDefault();

            setTimeout(() => slashPressed = false, 500);
            return;
        }

        if (slashPressed) {
            switch (event.key.toLowerCase()) {
                case 's':
                    document.getElementById('start_btn')?.click();
                    break;
                case 'r':
                    document.getElementById('resume_btn')?.click();
                    break;
                case 'x':
                    document.getElementById('stop_btn')?.click();
                    break;
                case 'a':
                    document.getElementById('abort_btn')?.click();
                    break;
                case 'z':
                    document.getElementById('step_btn')?.click();
                    break;
                case 't':
                    document.getElementById('target-btn')?.click();
                    break;
            }

            slashPressed = false;
        }
    });
