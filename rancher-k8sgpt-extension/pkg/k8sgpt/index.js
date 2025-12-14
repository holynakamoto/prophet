// Main plugin export for Rancher UI Extensions
import { init } from './product';

// Import components so they get bundled
import ListComponent from './list/core.k8sgpt.ai.result.vue';
import DetailComponent from './detail/core.k8sgpt.ai.result.vue';
import SeverityBadge from './components/SeverityBadge.vue';
import model from './models/core.k8sgpt.ai.result';

// Export the init function as default (Rancher expects this)
export default function(plugin) {
  // Register the product
  init(plugin, plugin.store);
  
  // Register components
  plugin.register('list', 'core.k8sgpt.ai.result', ListComponent);
  plugin.register('detail', 'core.k8sgpt.ai.result', DetailComponent);
  plugin.register('component', 'SeverityBadge', SeverityBadge);
  plugin.register('model', 'core.k8sgpt.ai.result', model);
}

// Provide plugin metadata
export const metadata = {
  name:        'k8sgpt',
  description: 'K8sGPT Diagnostics - View AI-powered Kubernetes troubleshooting results'
};
