import { CreateTopicInput } from '@/data/topics/validation'
import { supabase } from '@/clients/supabase.server'

export async function getTopicCountByApplicationId(applicationId: string) {
  const { count, error } = await supabase
    .from('topics')
    .select('*', { count: 'exact', head: true })
    .eq('application_id', applicationId)

  if (error) throw error
  return count
}

export async function createTopic(input: CreateTopicInput) {
  const { data, error } = await supabase
    .from('topics')
    .insert(input)
    .select('id, name, description, created_at, updated_at')
    .single()

  if (error) throw error
  return data
}
