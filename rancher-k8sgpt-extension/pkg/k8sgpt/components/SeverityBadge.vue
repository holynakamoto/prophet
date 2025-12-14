<script>
export default {
  name: 'SeverityBadge',

  props: {
    result: {
      type:     Object,
      required: true,
    },
  },

  computed: {
    severity() {
      const errors = (this.result && this.result.spec && this.result.spec.error) || [];
      if (errors.length === 0) {
        return 'info';
      }
      const text = errors.map(e => (e.text || '').toLowerCase()).join(' ');
      if (text.includes('error') || text.includes('crash') || text.includes('fail')) {
        return 'error';
      }
      if (text.includes('warn') || text.includes('pending') || text.includes('not ready')) {
        return 'warning';
      }
      return 'info';
    },

    severityClass() {
      return `severity-${ this.severity }`;
    },

    badgeColor() {
      switch (this.severity) {
      case 'error': return '#dc3545';
      case 'warning': return '#fd7e14';
      case 'info': return '#28a745';
      default: return '#6c757d';
      }
    },

    label() {
      return this.severity.charAt(0).toUpperCase() + this.severity.slice(1);
    },
  },
};
</script>

<template>
  <span
    class="severity-badge"
    :class="severityClass"
    :style="{ backgroundColor: badgeColor }"
  >
    {{ label }}
  </span>
</template>

<style lang="scss" scoped>
.severity-badge {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
  color: white;
  text-transform: uppercase;

  &.severity-error {
    background-color: #dc3545;
  }

  &.severity-warning {
    background-color: #fd7e14;
  }

  &.severity-info {
    background-color: #28a745;
  }
}
</style>
