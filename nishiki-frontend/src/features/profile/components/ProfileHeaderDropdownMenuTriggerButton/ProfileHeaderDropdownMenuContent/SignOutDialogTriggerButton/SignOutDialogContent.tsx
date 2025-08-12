import {
  Button,
  DialogBody,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui';
import { useAuth } from '@/contexts/AuthContext';

import { useState } from 'react';

interface ISignOutDialogContentProps {
  /**
   * The function to close the parent UI component.
   */
  onParentClose: () => void;
}

export const SignOutDialogContent = ({ onParentClose }: ISignOutDialogContentProps) => {
  const [isLoading, setIsLoading] = useState(false);
  const { logout } = useAuth();

  /**
   * Handle the cancel button click.
   * It closes the parent UI component.
   */
  const handleCancel = () => {
    onParentClose();
  };

  /**
   * Handle the SignOut button click.
   * Trigger the logout via Authentik.
   * @returns void
   */
  const handleSignOut = async () => {
    if (isLoading) return;
    setIsLoading(true);

    try {
      await logout();
      onParentClose();
    } catch (error) {
      console.error('Logout failed:', error);
      // Show error to user or handle gracefully
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <DialogContent>
      <DialogHeader>
        <DialogTitle>Sign out</DialogTitle>
      </DialogHeader>
      <DialogBody>
        <p>Are you sure you want to sign out?</p>
      </DialogBody>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="cancel" size="sm" onClick={handleCancel}>
            Cancel
          </Button>
        </DialogClose>
        <Button variant="danger-outline" size="sm" onClick={handleSignOut} disabled={isLoading}>
          Sign out
        </Button>
      </DialogFooter>
    </DialogContent>
  );
};
