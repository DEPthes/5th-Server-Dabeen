<template>
  <div class="page">
    <header class="header">
      <h1>📝 TODO</h1>
      <span class="subtitle">Docker Swarm 실습</span>
    </header>

    <!-- 입력 영역 -->
    <form class="input-form" @submit.prevent="addTodo">
      <input
        v-model="newTitle"
        class="input"
        placeholder="새 할 일을 입력하세요..."
        :disabled="submitting"
        maxlength="255"
      />
      <button type="submit" class="btn btn-add" :disabled="submitting || !newTitle.trim()">
        {{ submitting ? '...' : '추가' }}
      </button>
    </form>

    <!-- 상태 표시 -->
    <div v-if="error" class="alert alert-error">{{ error }}</div>
    <div v-if="loading" class="loading">불러오는 중...</div>

    <!-- 통계 -->
    <div v-if="!loading && todos.length" class="stats">
      전체 {{ todos.length }}개 · 완료 {{ doneCount }}개 · 미완료 {{ todos.length - doneCount }}개
    </div>

    <!-- TODO 목록 -->
    <ul v-if="!loading" class="todo-list">
      <li
        v-for="todo in todos"
        :key="todo.id"
        class="todo-item"
        :class="{ done: todo.done }"
      >
        <button
          class="check-btn"
          :title="todo.done ? '미완료로 변경' : '완료로 변경'"
          @click="toggleTodo(todo)"
        >
          {{ todo.done ? '✅' : '⬜' }}
        </button>
        <span class="todo-title">{{ todo.title }}</span>
        <span class="todo-date">{{ formatDate(todo.created_at) }}</span>
        <button class="btn btn-delete" @click="removeTodo(todo.id)">삭제</button>
      </li>
    </ul>

    <div v-if="!loading && !todos.length" class="empty">
      할 일이 없습니다. 위에서 추가해보세요!
    </div>
  </div>
</template>

<script setup lang="ts">
interface Todo {
  id: number
  title: string
  done: boolean
  created_at: string
  updated_at: string
}

const config = useRuntimeConfig()
const base = config.public.apiBase

const todos = ref<Todo[]>([])
const newTitle = ref('')
const loading = ref(false)
const submitting = ref(false)
const error = ref('')

const doneCount = computed(() => todos.value.filter(t => t.done).length)

const fetchTodos = async () => {
  loading.value = true
  error.value = ''
  try {
    todos.value = await $fetch<Todo[]>(`${base}/todos`)
  } catch (e: any) {
    error.value = `목록을 불러올 수 없습니다: ${e?.message ?? e}`
  } finally {
    loading.value = false
  }
}

const addTodo = async () => {
  const title = newTitle.value.trim()
  if (!title) return
  submitting.value = true
  error.value = ''
  try {
    await $fetch(`${base}/todos`, {
      method: 'POST',
      body: { title },
    })
    newTitle.value = ''
    await fetchTodos()
  } catch (e: any) {
    error.value = `추가 실패: ${e?.message ?? e}`
  } finally {
    submitting.value = false
  }
}

const toggleTodo = async (todo: Todo) => {
  try {
    await $fetch(`${base}/todos/${todo.id}`, {
      method: 'PUT',
      body: { done: !todo.done },
    })
    await fetchTodos()
  } catch (e: any) {
    error.value = `수정 실패: ${e?.message ?? e}`
  }
}

const removeTodo = async (id: number) => {
  try {
    await $fetch(`${base}/todos/${id}`, { method: 'DELETE' })
    await fetchTodos()
  } catch (e: any) {
    error.value = `삭제 실패: ${e?.message ?? e}`
  }
}

const formatDate = (iso: string) =>
  new Date(iso).toLocaleDateString('ko-KR', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })

onMounted(fetchTodos)
</script>

<style scoped>
.page        { max-width: 680px; margin: 40px auto; padding: 0 16px; font-family: 'Segoe UI', sans-serif; }
.header      { text-align: center; margin-bottom: 28px; }
.header h1   { font-size: 2rem; margin: 0; }
.subtitle    { color: #888; font-size: .9rem; }

.input-form  { display: flex; gap: 8px; margin-bottom: 16px; }
.input       { flex: 1; padding: 10px 14px; border: 1px solid #ddd; border-radius: 8px; font-size: 1rem; outline: none; }
.input:focus { border-color: #4f8ef7; box-shadow: 0 0 0 2px #4f8ef720; }

.btn         { padding: 10px 18px; border: none; border-radius: 8px; cursor: pointer; font-size: .95rem; transition: opacity .15s; }
.btn:disabled { opacity: .5; cursor: not-allowed; }
.btn-add     { background: #4f8ef7; color: #fff; }
.btn-add:hover:not(:disabled) { background: #3a7de0; }
.btn-delete  { background: #f76060; color: #fff; padding: 6px 12px; font-size: .85rem; }
.btn-delete:hover { background: #d94f4f; }

.alert-error { background: #fff0f0; border: 1px solid #f9b; color: #c00; padding: 10px 14px; border-radius: 8px; margin-bottom: 12px; }
.loading     { text-align: center; color: #aaa; padding: 20px; }
.stats       { color: #888; font-size: .85rem; margin-bottom: 12px; }

.todo-list   { list-style: none; padding: 0; margin: 0; }
.todo-item   { display: flex; align-items: center; gap: 10px; padding: 12px 8px; border-bottom: 1px solid #f0f0f0; transition: background .1s; }
.todo-item:hover { background: #fafafa; }
.todo-item.done .todo-title { text-decoration: line-through; color: #bbb; }

.check-btn   { background: none; border: none; cursor: pointer; font-size: 1.2rem; padding: 0; }
.todo-title  { flex: 1; font-size: 1rem; }
.todo-date   { font-size: .75rem; color: #bbb; white-space: nowrap; }

.empty       { text-align: center; color: #ccc; padding: 40px; font-size: 1.1rem; }
</style>
