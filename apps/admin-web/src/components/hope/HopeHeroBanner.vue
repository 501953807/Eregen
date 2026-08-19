<template>
  <div class="hero-banner">
    <div class="hero-content">
      <div class="hero-text">
        <h1>Hello Admin!</h1>
        <p>欢迎回来，今天也是元气满满的一天~</p>
      </div>
      <button class="announcements-btn">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z"/>
        </svg>
        Announcements
      </button>
    </div>
    <div class="hero-decoration">
      <div class="hero-circle hero-circle--1"></div>
      <div class="hero-circle hero-circle--2"></div>
      <div class="hero-circle hero-circle--3"></div>
    </div>
  </div>
</template>

<style scoped>
/* ─── Hope UI Hero Banner ─── */
.hero-banner {
  position: relative;
  height: 264px;
  background: var(--hope-primary-gradient);
  overflow: hidden;
  border-radius: 20px 20px 0 0;
  margin: 1.5rem;
  margin-bottom: 0;
}

.hero-content {
  position: relative;
  z-index: 1;
  padding: 3rem;
  color: white;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.hero-text h1 {
  font-size: 2.25rem;
  font-weight: 700;
  margin-bottom: 0.25rem;
  letter-spacing: -0.02em;
  color: #ffffff;
}

.hero-text p {
  font-size: 1rem;
  opacity: 0.85;
  font-weight: 400;
  color: #ffffff;
}

.hero-actions {
  flex-shrink: 0;
}

.announcements-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 0.5rem 1rem;
  background: rgba(255, 255, 255, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.28);
  border-radius: 6px;
  color: white;
  font-size: 0.8125rem;
  font-weight: 500;
  cursor: pointer;
  backdrop-filter: blur(10px);
  transition: background 0.15s ease, border-color 0.15s ease;
}

.announcements-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  border-color: rgba(255, 255, 255, 0.45);
}

/* ─── Decorative circles: Hope UI floating shapes ─── */
.hero-decoration {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
}

.hero-circle {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.04);
}

.hero-circle--1 {
  bottom: -80px;
  right: -40px;
  width: 240px;
  height: 240px;
}

.hero-circle--2 {
  top: -40px;
  right: 120px;
  width: 120px;
  height: 120px;
  background: rgba(255, 255, 255, 0.04);
}

.hero-circle--3 {
  bottom: 30px;
  right: 200px;
  width: 60px;
  height: 60px;
  background: rgba(255, 255, 255, 0.05);
}
</style>
