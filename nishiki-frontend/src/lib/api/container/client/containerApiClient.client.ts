import { request } from '@/lib/api/common/client';
import { IContainer, IGroup } from '@/types/definition';

import { Err, Ok, Result } from 'result-ts-type';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || '';

export interface IPostContainerRequestBody {
  /**
   * An identifier of a group that a new container will belong to
   */
  groupId: IGroup['id'];
  /**
   * New container name which a user create
   */
  name: string;
}

export interface IPostContainerResponse {
  /**
   * An identifier of a newly created container
   */
  containerId: IContainer['id'];
}

/**
 * Function to send a request to the API to create a new container
 * @param requestBody - A {@link IPostContainerRequestBody} object to be sent to API as the request body
 * @returns A {@link IPostContainerResponse} object for success, an error message if fails
 */
export const postCreateContainer = async (
  requestBody: IPostContainerRequestBody,
): Promise<Result<IPostContainerResponse, string>> => {
  try {
    const res = await request<IPostContainerResponse>({
      url: `${API_BASE_URL}/containers`,
      method: 'POST',
      options: {
        body: JSON.stringify(requestBody),
      },
    });
    return Ok({ containerId: res.containerId });
  } catch (err) {
    if (err instanceof Error) {
      return Err(err.message);
    }
    return Err('API response is invalid');
  }
};

export interface IPutRenameContainerRequestBody {
  /**
   * a new container name which a user input
   */
  containerName: string;
}

/**
 * Function to send a request to the API to rename the container
 * @param containerId - The identifier of container whose name a user is willing to change
 * @param requestBody - A {@link IPutRenameContainerRequestBody} object to be sent to API as the request body
 * @returns undefined for success, an error message if fails
 */
export const putRenameContainer = async (
  containerId: IContainer['id'],
  requestBody: IPutRenameContainerRequestBody,
): Promise<Result<undefined, string>> => {
  try {
    await request({
      url: `${API_BASE_URL}/containers/${containerId}`,
      method: 'PUT',
      options: { body: JSON.stringify(requestBody) },
    });
    return Ok(undefined);
  } catch (err) {
    if (err instanceof Error) {
      return Err(err.message);
    }
    return Err('API response is invalid');
  }
};

/**
 * Function to send a request to API to delete container
 * @param containerId - The identifier of container which a user is willing to delete
 * @returns undefined for success, an error message if fails
 */
export const deleteContainer = async (
  containerId: IContainer['id'],
): Promise<Result<undefined, string>> => {
  try {
    await request({
      url: `${API_BASE_URL}/containers/${containerId}`,
      method: 'DELETE',
    });
    return Ok(undefined);
  } catch (err) {
    if (err instanceof Error) {
      return Err(err.message);
    }
    return Err('API response is invalid');
  }
};

/**
 * Fetch containers for a specific group.
 * @param groupId - The ID of the group
 * @returns Array of IContainer objects
 */
export const getContainersByGroup = async (
  groupId: string,
): Promise<Result<IContainer[], string>> => {
  try {
    const data: IContainer[] = await request<IContainer[]>({
      url: `${API_BASE_URL}/groups/${groupId}/containers`,
      method: 'GET',
    });
    return Ok(data);
  } catch (err) {
    return Err(`Failed to fetch containers: ${err instanceof Error ? err.message : String(err)}`);
  }
};
