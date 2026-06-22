(function() {
    const el = document.querySelector('[data-post-id]');
    if (!el) return;

    const postId = el.getAttribute('data-post-id');
    let activeSeconds = 0;
    let lastInteraction = Date.now();

    function resetIdle() {
        lastInteraction = Date.now();
    }

    window.addEventListener('mousemove', resetIdle);
    window.addEventListener('keypress', resetIdle);
    window.addEventListener('scroll', resetIdle);
    window.addEventListener('click', resetIdle);

    setInterval(() => {
        const now = Date.now();
        const isTabActive = document.visibilityState === 'visible' && document.hasFocus();
        const isNotIdle = (now - lastInteraction) < 45000;

        if (isTabActive && isNotIdle) {
            activeSeconds++;
        }

        if (activeSeconds >= 15) {
            sendHeartbeat(15);
            activeSeconds = 0;
        }
    }, 1000);

    function sendRemaining() {
        if (activeSeconds > 0) {
            sendHeartbeat(activeSeconds);
            activeSeconds = 0;
        }
    }

    document.addEventListener('visibilitychange', () => {
        if (document.visibilityState === 'hidden') {
            sendRemaining();
        }
    });
    window.addEventListener('pagehide', sendRemaining);

    function sendHeartbeat(seconds) {
        const url = `/api/v1/posts/${postId}/read-time`;
        const payload = JSON.stringify({ seconds: seconds });

        if (navigator.sendBeacon) {
            const blob = new Blob([payload], { type: 'application/json' });
            navigator.sendBeacon(url, blob);
        } else {
            fetch(url, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: payload
            }).catch(() => {});
        }
    }
})();