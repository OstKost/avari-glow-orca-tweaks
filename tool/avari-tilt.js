(() => {
  'use strict';
  const selector = '.terminal-tab-strip [data-tab-id][data-active], [data-worktree-card-surface][data-worktree-card-active="primary"], [data-native-chat-scroll] .bg-chat-user-surface';
  const motion = matchMedia('(prefers-reduced-motion: no-preference) and (hover: hover) and (pointer: fine)');
  let current = null, frame = 0, point = null;
  function reset() {
    if (frame) cancelAnimationFrame(frame);
    frame = 0; point = null;
    if (current) { current.style.removeProperty('--avari-tilt-axis'); current.style.removeProperty('--avari-tilt-angle'); current = null; }
  }
  document.addEventListener('pointermove', event => {
    if (!motion.matches || event.pointerType !== 'mouse' || event.buttons) { reset(); return; }
    const target = event.target instanceof Element ? event.target.closest(selector) : null;
    if (!target) { reset(); return; }
    if (current !== target) { reset(); current = target; }
    point = {x:event.clientX, y:event.clientY};
    if (frame) return;
    frame = requestAnimationFrame(() => {
      frame = 0;
      if (!current || !point || !current.isConnected) { reset(); return; }
      const r = current.getBoundingClientRect();
      if (!r.width || !r.height) return;
      const x = Math.max(-1, Math.min(1, (point.x-r.left)/r.width*2-1));
      const y = Math.max(-1, Math.min(1, (point.y-r.top)/r.height*2-1));
      const magnitude = Math.min(1, Math.hypot(x,y));
      current.style.setProperty('--avari-tilt-axis', `${(-y).toFixed(3)} ${x.toFixed(3)} 0`);
      const maxAngle = current.hasAttribute('data-tab-id') || current.classList.contains('bg-chat-user-surface') ? 1 : 3;
      current.style.setProperty('--avari-tilt-angle', `${(magnitude * maxAngle).toFixed(2)}deg`);
    });
  }, {passive:true});
  document.addEventListener('pointerdown', reset, true);
  document.addEventListener('pointercancel', reset, true);
  document.addEventListener('pointerout', event => { if (current && !current.contains(event.relatedTarget)) reset(); }, true);
  // Paint a non-interactive light over the actual sidebar bounds, without
  // changing their stacking contexts, scroll behaviour or hit testing.
  const sidebarSelector = '.bg-worktree-sidebar.flex.flex-col.overflow-hidden:has([data-worktree-sidebar-container]), .bg-sidebar.flex.flex-col.flex-1.min-w-0.overflow-hidden';
  let light = null, lightPanel = null, lightFrame = 0, lightPoint = null;
  function hideLight() {
    if (lightFrame) cancelAnimationFrame(lightFrame);
    lightFrame = 0; lightPoint = null; lightPanel = null;
    if (light) light.classList.remove('avari-spotlight-visible');
  }
  document.addEventListener('pointermove', event => {
    if (!motion.matches || event.pointerType !== 'mouse' || event.buttons) { hideLight(); return; }
    const panel = event.target instanceof Element ? event.target.closest(sidebarSelector) : null;
    if (!panel) { hideLight(); return; }
    if (!light) { light = document.createElement('div'); light.className = 'avari-sidebar-spotlight'; light.setAttribute('aria-hidden', 'true'); document.body.appendChild(light); }
    lightPanel = panel; lightPoint = {x:event.clientX, y:event.clientY};
    if (lightFrame) return;
    lightFrame = requestAnimationFrame(() => {
      lightFrame = 0;
      if (!lightPanel || !lightPanel.isConnected || !lightPoint) { hideLight(); return; }
      const r = lightPanel.getBoundingClientRect();
      Object.assign(light.style, {left:`${r.left}px`, top:`${r.top}px`, width:`${r.width}px`, height:`${r.height}px`});
      light.style.setProperty('--avari-spot-x', `${lightPoint.x-r.left}px`);
      light.style.setProperty('--avari-spot-y', `${lightPoint.y-r.top}px`);
      light.classList.add('avari-spotlight-visible');
    });
  }, {passive:true});
  document.addEventListener('pointerdown', hideLight, true);
  document.addEventListener('pointercancel', hideLight, true);
  document.addEventListener('pointerout', event => { if (lightPanel && !lightPanel.contains(event.relatedTarget)) hideLight(); }, true);
  motion.addEventListener('change', () => { if (!motion.matches) hideLight(); });
  document.addEventListener('visibilitychange', () => { if (document.hidden) { reset(); hideLight(); } });
  motion.addEventListener('change', () => { if (!motion.matches) reset(); });
})();
