// The chosen chart style is external state (localStorage + an attribute on
// <html>), so it's exposed as a store rather than restored inside an effect.
const KEY = "shapehill:style";

let listeners: Array<() => void> = [];
let current: string | null = null;

export function getStyle() {
  if (current === null) {
    current = (typeof window !== "undefined" && window.localStorage.getItem(KEY)) || "";
  }
  return current;
}

// The server has no localStorage; it always renders the default theme.
export function getServerStyle() {
  return "";
}

export function setStyle(next: string) {
  current = next;
  window.localStorage.setItem(KEY, next);
  listeners.forEach((l) => l());
}

export function subscribe(listener: () => void) {
  listeners.push(listener);
  return () => {
    listeners = listeners.filter((l) => l !== listener);
  };
}
