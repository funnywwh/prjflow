<template>
  <div class="markdown-editor">
    <a-tabs v-model:activeKey="activeTab" v-if="!readonly">
      <a-tab-pane key="edit" tab="编辑">
        <a-textarea
          :value="modelValue || ''"
          @update:value="handleInput"
          :placeholder="placeholder"
          :rows="rows"
          style="font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace"
        />
      </a-tab-pane>
      <a-tab-pane key="preview" tab="预览">
        <div class="markdown-preview" v-html="renderedMarkdown"></div>
      </a-tab-pane>
    </a-tabs>
    <div v-else class="markdown-preview" v-html="renderedMarkdown"></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { marked } from 'marked'
import hljs from 'highlight.js'
import 'highlight.js/styles/github.css'

interface Props {
  modelValue?: string
  placeholder?: string
  rows?: number
  readonly?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: '请输入Markdown内容...',
  rows: 8,
  readonly: false
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const activeTab = ref('edit')

// 配置marked
marked.setOptions({
  highlight: function(code: string, lang: string) {
    if (lang && hljs.getLanguage(lang)) {
      try {
        return hljs.highlight(code, { language: lang }).value
      } catch (err) {
        console.error('Highlight error:', err)
      }
    }
    return hljs.highlightAuto(code).value
  },
  breaks: true,
  gfm: true
} as any)

// 渲染Markdown
const renderedMarkdown = computed(() => {
  if (!props.modelValue || props.modelValue.trim() === '') {
    return '<p class="empty-text">暂无内容</p>'
  }
  return marked.parse(props.modelValue)
})

// 处理输入
const handleInput = (value: string) => {
  emit('update:modelValue', value)
}

// 监听modelValue变化，自动切换到预览
watch(() => props.modelValue, () => {
  if (activeTab.value === 'preview' && props.modelValue) {
    // 如果正在预览且有内容，保持预览状态
  }
})
</script>

<style scoped>
.markdown-editor {
  width: 100%;
}

.markdown-preview {
  min-height: 200px;
  padding: 12px;
  border: 1px solid #d9d9d9;
  border-radius: 4px;
  background: #fff;
  overflow-y: auto;
  max-height: 600px;
}

.markdown-preview :deep(h1),
.markdown-preview :deep(h2),
.markdown-preview :deep(h3),
.markdown-preview :deep(h4),
.markdown-preview :deep(h5),
.markdown-preview :deep(h6) {
  margin-top: 16px;
  margin-bottom: 8px;
  font-weight: 600;
  line-height: 1.25;
}

.markdown-preview :deep(h1) {
  font-size: 2em;
  border-bottom: 1px solid #eaecef;
  padding-bottom: 0.3em;
}

.markdown-preview :deep(h2) {
  font-size: 1.5em;
  border-bottom: 1px solid #eaecef;
  padding-bottom: 0.3em;
}

.markdown-preview :deep(h3) {
  font-size: 1.25em;
}

.markdown-preview :deep(p) {
  margin-bottom: 16px;
  line-height: 1.6;
}

.markdown-preview :deep(ul),
.markdown-preview :deep(ol) {
  margin-bottom: 16px;
  padding-left: 2em;
}

.markdown-preview :deep(li) {
  margin-bottom: 4px;
}

.markdown-preview :deep(blockquote) {
  padding: 0 1em;
  color: #6a737d;
  border-left: 0.25em solid #dfe2e5;
  margin-bottom: 16px;
}

.markdown-preview :deep(code) {
  padding: 0.2em 0.4em;
  margin: 0;
  font-size: 85%;
  background-color: rgba(27, 31, 35, 0.05);
  border-radius: 3px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}

.markdown-preview :deep(pre) {
  padding: 16px;
  overflow: auto;
  font-size: 85%;
  line-height: 1.45;
  background-color: #f6f8fa;
  border-radius: 6px;
  margin-bottom: 16px;
}

.markdown-preview :deep(pre code) {
  display: inline;
  max-width: auto;
  padding: 0;
  margin: 0;
  overflow: visible;
  line-height: inherit;
  word-wrap: normal;
  background-color: transparent;
  border: 0;
}

.markdown-preview :deep(table) {
  border-collapse: collapse;
  margin-bottom: 16px;
  width: 100%;
}

.markdown-preview :deep(table th),
.markdown-preview :deep(table td) {
  padding: 6px 13px;
  border: 1px solid #dfe2e5;
}

.markdown-preview :deep(table th) {
  font-weight: 600;
  background-color: #f6f8fa;
}

.markdown-preview :deep(table tr:nth-child(2n)) {
  background-color: #f6f8fa;
}

.markdown-preview :deep(a) {
  color: #0366d6;
  text-decoration: none;
}

.markdown-preview :deep(a:hover) {
  text-decoration: underline;
}

.markdown-preview :deep(img) {
  max-width: 100%;
  box-sizing: content-box;
  background-color: #fff;
}

.markdown-preview :deep(hr) {
  height: 0.25em;
  padding: 0;
  margin: 24px 0;
  background-color: #e1e4e8;
  border: 0;
}

.empty-text {
  color: #999;
  font-style: italic;
}
</style>

