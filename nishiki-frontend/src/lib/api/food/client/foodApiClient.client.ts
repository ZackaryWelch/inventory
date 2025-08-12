import { request } from '@/lib/api/common/client';
import { IFood } from '@/types/definition';

import { Err, Ok, Result } from 'result-ts-type';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || '';

/**
 * Interface for creating a new food item.
 */
export interface ICreateFoodRequest {
  name: string;
  quantity: number | null;
  category: string;
  unit: string | null;
  expiry: string | null; // ISO date string
  containerId: string;
}

/**
 * Interface for updating a food item.
 */
export interface IUpdateFoodRequest {
  name?: string;
  quantity?: number | null;
  category?: string;
  unit?: string | null;
  expiry?: string | null; // ISO date string
}

/**
 * Add a new food item to a container.
 * @param foodData - The food data to create
 * @returns The created food object
 */
export const createFood = async (foodData: ICreateFoodRequest): Promise<Result<IFood, string>> => {
  try {
    const data: IFood = await request<IFood>({
      url: API_BASE_URL + '/foods',
      method: 'POST',
      options: {
        body: JSON.stringify(foodData),
      },
    });
    return Ok(data);
  } catch (err) {
    return Err(`Failed to create food: ${err instanceof Error ? err.message : String(err)}`);
  }
};

/**
 * Update a food item.
 * @param foodId - The ID of the food to update
 * @param foodData - The updated food data
 * @returns The updated food object
 */
export const updateFood = async (
  foodId: string,
  foodData: IUpdateFoodRequest,
): Promise<Result<IFood, string>> => {
  try {
    const data: IFood = await request<IFood>({
      url: `${API_BASE_URL}/foods/${foodId}`,
      method: 'PUT',
      options: {
        body: JSON.stringify(foodData),
      },
    });
    return Ok(data);
  } catch (err) {
    return Err(`Failed to update food: ${err instanceof Error ? err.message : String(err)}`);
  }
};

/**
 * Delete a food item.
 * @param foodId - The ID of the food to delete
 * @returns Success or error result
 */
export const deleteFood = async (foodId: string): Promise<Result<void, string>> => {
  try {
    await request<void>({
      url: `${API_BASE_URL}/foods/${foodId}`,
      method: 'DELETE',
    });
    return Ok(undefined);
  } catch (err) {
    return Err(`Failed to delete food: ${err instanceof Error ? err.message : String(err)}`);
  }
};
