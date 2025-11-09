import { CreateApplicationInput } from '@/data/apps/validation'
import { supabase } from '@/clients/supabase.server'

export async function createApplication(input: CreateApplicationInput) {
  const { data, error } = await supabase
    .from('applications')
    .insert(input)
    .select('id, name, description, created_at, updated_at')
    .single()

  if (error) throw error
  return data
}

export async function getApplicationsWithTopicCountByUserId(userId: string) {
  const { data, error } = await supabase
    .from('applications')
    .select('*, topics(count)')
    .eq('user_id', userId)
    .order('created_at', { ascending: false })

  if (error) throw error
  return data
}
