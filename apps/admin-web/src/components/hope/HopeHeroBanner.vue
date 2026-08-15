<template>
  <div class="hero-banner">
    <div class="hero-content">
      <div class="hero-text">
        <h1>Hello Admin!</h1>
        <p>欢迎回来，今天也是元气满满的一天~</p>
      </div>
      <button class="announcements-btn">
        Announcements
      </button>
    </div>
    <div class="hero-decoration"></div>
  </div>
</template>

<style scoped>
.hero-banner {
  position: relative;
  height: 264px;
  background: var(--hope-primary-gradient);
  overflow: hidden;
  border-radius: var(--hope-radius-lg) var(--hope-radius-lg) 0 0;
  margin: 1.5rem;
  margin-bottom: 0;
}

.hero-content {
  position: relative;
  z-index: 1;
  padding: 3rem;
  color: white;
}

.hero-text h1 {
  font-size: 2.25rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.hero-text p {
  font-size: 0.875rem;
  opacity: 0.8;
}

.announcements-btn {
  position: absolute;
  top: 2rem;
  right: 2rem;
  padding: 0.5rem 1rem;
  background: rgba(255,255,255,0.2);
  border: 1px solid rgba(255,255,255,0.3);
  border-radius: var(--hope-radius-sm);
  color: white;
  font-size: 0.8125rem;
  cursor: pointer;
  backdrop-filter: blur(10px);
}

.announcements-btn:hover {
  background: rgba(255,255,255,0.3);
}

.hero-decoration {
  position: absolute;
  bottom: -50px;
  right: -50px;
  width: 300px;
  height: 300px;
  background: rgba(255,255,255,0.1);
  border-radius: 50%;
}

.hero-decoration::before {
  content: '';
  position: absolute;
  top: 50px;
  left: 50px;
  width: 200px;
  height: 200px;
  background: rgba(255,255,255,0.05);
  border-radius: 50%;
}
</style>
