// src/components/common/EmptyState.js
import React from 'react';
import { Box, Typography, Button } from '@mui/material';
import { Link as RouterLink } from 'react-router-dom';

const EmptyState = ({ icon, title, description, actionText, actionTo }) => {
  return (
    <Box
      sx={{
        textAlign: 'center',
        p: 4,
        border: '1px dashed',
        borderColor: 'divider',
        borderRadius: 1,
        backgroundColor: 'action.hover',
      }}
    >
      {icon && <Box sx={{ fontSize: 60, color: 'text.secondary', mb: 2 }}>{icon}</Box>}
      <Typography variant="h6" component="h3" gutterBottom>
        {title}
      </Typography>
      <Typography color="text.secondary" sx={{ mb: 3 }}>
        {description}
      </Typography>
      {actionText && actionTo && (
        <Button
          variant="contained"
          component={RouterLink}
          to={actionTo}
        >
          {actionText}
        </Button>
      )}
    </Box>
  );
};

export default EmptyState;
