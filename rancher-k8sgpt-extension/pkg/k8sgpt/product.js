export function init(plugin, store) {
  const {
    product,
    basicType,
    virtualType,
  } = plugin.DSL(store, 'k8sgpt');

  // Define the product (appears in left sidebar)
  product({
    icon:                  'globe',
    inStore:               'cluster',
    weight:                100,
    to:                    { name: 'c-cluster-product-resource', params: { product: 'k8sgpt', resource: 'core.k8sgpt.ai.result' } }
  });

  // Virtual type for the K8sGPT Results
  virtualType({
    label:        'K8sGPT Results',
    name:         'core.k8sgpt.ai.result',
    namespaced:   true,
    route:        { name: 'c-cluster-product-resource', params: { product: 'k8sgpt', resource: 'core.k8sgpt.ai.result' } },
    weight:       100,
  });

  // Basic type registration
  basicType(['core.k8sgpt.ai.result'], 'k8sgpt');
}

