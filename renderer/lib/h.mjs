// Minimal JSX-like helper for Satori-compatible virtual nodes.
export function h(type, props, ...children) {
  const flatChildren = [];
  const flatten = (child) => {
    if (Array.isArray(child)) {
      for (const nested of child) flatten(nested);
    } else if (child != null && child !== false) {
      flatChildren.push(child);
    }
  };
  for (const child of children) flatten(child);

  return {
    type,
    props: {
      ...(props ?? {}),
      children: flatChildren,
    },
  };
}
