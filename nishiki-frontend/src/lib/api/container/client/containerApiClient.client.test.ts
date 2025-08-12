import { request } from '@/lib/api/common/client';

// Target functions to test
import {
  deleteContainer,
  postCreateContainer,
  putRenameContainer,
} from './containerApiClient.client';

// Mock request function
jest.mock('@/lib/api/common/client', () => ({
  request: jest.fn(),
}));

describe('containerApiClient', () => {
  // Clear mocks after each test
  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('Containers', () => {
    // Arrange mock data for containers
    const mockContainerId = 'container1';
    //test for creating new container function
    describe('postCreateContainer', () => {
      /* Arrange common mock data */
      const mockRequestBody = { groupId: 'groupId', name: 'newContainer' };

      it('should return OK result with new container ID', async () => {
        /* Arrange */
        const mockResponse = { containerId: mockContainerId };
        (request as jest.Mock).mockResolvedValue(mockResponse);

        /* Act */
        const result = await postCreateContainer(mockRequestBody);

        /* Assert*/
        expect(result.unwrap()).toEqual(mockResponse);
        expect(request).toHaveBeenCalledWith({
          url: expect.stringContaining(`/containers`),
          method: 'POST',
          options: {
            body: JSON.stringify(mockRequestBody),
          },
        });
      });

      it('should return error result if API request fails', async () => {
        /* Arrange */
        const mockError = new Error('API error');
        (request as jest.Mock).mockRejectedValue(mockError);

        /* Act */
        const result = await postCreateContainer(mockRequestBody);

        /* Assert */
        expect(result.unwrapError()).toEqual(mockError.message);
      });
    });

    describe('putRenameContainer', () => {
      /* Arrange common mock data */
      const mockRequestBody = { containerName: 'newContainerName' };
      it('should return OK result with undefined', async () => {
        /* Arrange */
        (request as jest.Mock).mockResolvedValue({});

        /* Act */
        const result = await putRenameContainer(mockContainerId, mockRequestBody);

        /* Assert */
        expect(result.unwrap()).toBe(undefined);
        expect(request).toHaveBeenCalledWith({
          url: expect.stringContaining(`containers/${mockContainerId}`),
          method: 'PUT',
          options: {
            body: JSON.stringify(mockRequestBody),
          },
        });
      });

      it('should return error result if API request fails', async () => {
        /* Arrange */
        const mockError = new Error('API error');
        (request as jest.Mock).mockRejectedValue(mockError);

        /* Act */
        const result = await putRenameContainer(mockContainerId, mockRequestBody);

        /* Assert */
        expect(result.unwrapError()).toBe(mockError.message);
      });
    });

    describe('deleteContainer', () => {
      it('should return Ok result on success', async () => {
        /* Arrange */
        (request as jest.Mock).mockResolvedValue({});

        /* Act */
        const result = await deleteContainer(mockContainerId);

        /* Assert */
        expect(result.ok).toBeTruthy();
        expect(request).toHaveBeenCalledWith({
          url: expect.stringContaining(`/containers/${mockContainerId}`),
          method: 'DELETE',
        });
      });

      it('should return Err result if API request fails', async () => {
        /* Arrange */
        const mockError = new Error('API error');
        (request as jest.Mock).mockRejectedValue(mockError);

        /* Act */
        const result = await deleteContainer(mockContainerId);

        /* Assert */
        expect(result.err).toBeTruthy();
        expect(result.unwrapError()).toBe(mockError.message);
      });
    });
  });
});
