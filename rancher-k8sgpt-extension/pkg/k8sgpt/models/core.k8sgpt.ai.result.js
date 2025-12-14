import SteveModel from '@shell/plugins/steve/steve-class';

export default class K8sGPTResult extends SteveModel {
  get _availableActions() {
    const out = super._availableActions;

    return out;
  }

  // State color based on severity
  get stateColor() {
    const errors = (this.spec && this.spec.error) || [];

    if (errors.length === 0) {
      return 'success';
    }

    const text = errors.map(e => (e.text || '').toLowerCase()).join(' ');

    if (text.includes('error') || text.includes('crash') || text.includes('fail')) {
      return 'error';
    }
    if (text.includes('warn') || text.includes('pending')) {
      return 'warning';
    }

    return 'info';
  }

  // State display text
  get stateDisplay() {
    const errors = (this.spec && this.spec.error) || [];

    if (errors.length === 0) {
      return 'Healthy';
    }

    const text = errors.map(e => (e.text || '').toLowerCase()).join(' ');

    if (text.includes('error') || text.includes('crash') || text.includes('fail')) {
      return 'Error';
    }
    if (text.includes('warn') || text.includes('pending')) {
      return 'Warning';
    }

    return 'Info';
  }

  // Get explanation preview
  get explanationPreview() {
    const errors = (this.spec && this.spec.error) || [];

    if (errors.length === 0) {
      return 'No issues detected';
    }
    const text = errors.map(e => e.text || '').join(' | ');

    return text.length > 100 ? `${ text.substring(0, 100) }...` : text;
  }
}
