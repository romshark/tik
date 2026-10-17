// Mirror the system color scheme onto the data-theme attribute used by theme.css.
const darkScheme = matchMedia("(prefers-color-scheme: dark)");
const applyTheme = () => {
	document.documentElement.dataset.theme = darkScheme.matches ? "dark" : "light";
};
applyTheme();
darkScheme.addEventListener("change", applyTheme);
