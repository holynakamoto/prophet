<script>
import ResourceTable from '@shell/components/ResourceTable';
import SeverityBadge from '../components/SeverityBadge';

export default {
  name:       'K8sGPTResultList',
  components: { ResourceTable, SeverityBadge },

  props: {
    resource: {
      type:     String,
      required: true,
    },
    schema: {
      type:     Object,
      required: true,
    },
  },

  computed: {
    headers() {
      return [
        {
          name:          'state',
          labelKey:      'tableHeaders.state',
          value:         'stateDisplay',
          sort:          ['stateSort', 'nameSort'],
          width:         100,
          default:       'unknown',
          formatter:     'BadgeStateFormatter',
        },
        {
          name:      'name',
          labelKey:  'tableHeaders.name',
          value:     'nameDisplay',
          sort:      ['nameSort'],
          formatter: 'LinkDetail',
        },
        {
          name:     'namespace',
          labelKey: 'tableHeaders.namespace',
          value:    'namespace',
          sort:     ['namespace'],
        },
        {
          name:  'kind',
          label: 'Resource Kind',
          value: 'spec.kind',
          sort:  ['spec.kind'],
        },
        {
          name:  'involvedObject',
          label: 'Involved Object',
          value: 'spec.name',
          sort:  ['spec.name'],
        },
        {
          name:  'explanation',
          label: 'Explanation',
          value: 'explanationPreview',
        },
        {
          name:      'age',
          labelKey:  'tableHeaders.age',
          value:     'creationTimestamp',
          sort:      'creationTimestamp:desc',
          search:    false,
          formatter: 'LiveDate',
          width:     100,
        },
      ];
    },
  },
};
</script>

<template>
  <ResourceTable
    :schema="schema"
    :headers="headers"
    :namespaced="true"
    :group-by="null"
  />
</template>



