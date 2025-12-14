<script>
import CreateEditView from '@shell/mixins/CreateEditView';
import ResourceTabs from '@shell/components/form/ResourceTabs';
import Tab from '@shell/components/Tabbed/Tab';
import SeverityBadge from '../components/SeverityBadge';

export default {
  name:       'K8sGPTResultDetail',
  components: {
    ResourceTabs, Tab, SeverityBadge
  },
  mixins: [CreateEditView],

  props: {
    value: {
      type:     Object,
      required: true,
    },
  },

  computed: {
    errors() {
      return (this.value && this.value.spec && this.value.spec.error) || [];
    },

    involvedKind() {
      return (this.value && this.value.spec && this.value.spec.kind) || 'Unknown';
    },

    involvedName() {
      return (this.value && this.value.spec && this.value.spec.name) || '-';
    },

    parentObject() {
      return (this.value && this.value.spec && this.value.spec.parentObject) || '-';
    },

    details() {
      return (this.value && this.value.spec && this.value.spec.details) || '';
    },
  },
};
</script>

<template>
  <div class="k8sgpt-result-detail">
    <!-- Header -->
    <header class="result-header">
      <div class="header-left">
        <h1>{{ value.metadata.name }}</h1>
        <span class="namespace-badge">{{ value.metadata.namespace }}</span>
      </div>
      <SeverityBadge :result="value" />
    </header>

    <ResourceTabs
      v-model="value"
      mode="view"
    >
      <!-- Diagnostics Tab -->
      <Tab
        name="diagnostics"
        label="AI Diagnostics"
        :weight="100"
      >
        <div class="diagnostics-content">
          <h3>Detected Issues</h3>
          <div
            v-for="(error, index) in errors"
            :key="index"
            class="error-item"
          >
            <p class="error-text">
              {{ error.text }}
            </p>
            <div
              v-if="error.sensitive && error.sensitive.length > 0"
              class="sensitive-data"
            >
              <small>Contains sensitive data (masked)</small>
            </div>
          </div>
          <p
            v-if="errors.length === 0"
            class="no-errors"
          >
            No issues detected.
          </p>
        </div>
      </Tab>

      <!-- Resource Info Tab -->
      <Tab
        name="resource"
        label="Involved Resource"
        :weight="90"
      >
        <div class="resource-info">
          <div class="info-row">
            <span class="label">Kind:</span>
            <span class="value">{{ involvedKind }}</span>
          </div>
          <div class="info-row">
            <span class="label">Name:</span>
            <span class="value">{{ involvedName }}</span>
          </div>
          <div class="info-row">
            <span class="label">Namespace:</span>
            <span class="value">{{ value.metadata.namespace }}</span>
          </div>
          <div
            v-if="parentObject !== '-'"
            class="info-row"
          >
            <span class="label">Parent Object:</span>
            <span class="value">{{ parentObject }}</span>
          </div>
        </div>
      </Tab>

      <!-- Details Tab -->
      <Tab
        v-if="details"
        name="details"
        label="Details"
        :weight="80"
      >
        <pre class="details-text">{{ details }}</pre>
      </Tab>
    </ResourceTabs>
  </div>
</template>

<style lang="scss" scoped>
.k8sgpt-result-detail {
  .result-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    padding-bottom: 15px;
    border-bottom: 1px solid var(--border);

    .header-left {
      display: flex;
      align-items: center;
      gap: 12px;

      h1 {
        margin: 0;
        font-size: 20px;
        font-weight: 500;
      }

      .namespace-badge {
        padding: 4px 8px;
        background: var(--default-active-bg);
        border-radius: 4px;
        font-size: 12px;
        color: var(--muted);
      }
    }
  }

  .diagnostics-content {
    h3 {
      margin: 0 0 15px;
      font-size: 16px;
      font-weight: 500;
    }

    .error-item {
      padding: 15px;
      background: var(--body-bg);
      border: 1px solid var(--border);
      border-radius: 6px;
      margin-bottom: 12px;

      &:last-child {
        margin-bottom: 0;
      }

      .error-text {
        margin: 0;
        font-size: 14px;
        line-height: 1.6;
        white-space: pre-wrap;
        word-break: break-word;
      }

      .sensitive-data {
        margin-top: 10px;
        padding-top: 10px;
        border-top: 1px solid var(--border);
        color: var(--warning);
      }
    }

    .no-errors {
      color: var(--success);
      font-size: 14px;
      margin: 0;
    }
  }

  .resource-info {
    .info-row {
      display: flex;
      padding: 10px 0;
      border-bottom: 1px solid var(--border);
      font-size: 14px;

      &:last-child {
        border-bottom: none;
      }

      .label {
        width: 150px;
        color: var(--muted);
        font-weight: 500;
      }

      .value {
        flex: 1;
      }
    }
  }

  .details-text {
    background: var(--body-bg);
    padding: 15px;
    border: 1px solid var(--border);
    border-radius: 6px;
    font-size: 13px;
    font-family: monospace;
    overflow-x: auto;
    margin: 0;
    white-space: pre-wrap;
    word-break: break-word;
  }
}
</style>
