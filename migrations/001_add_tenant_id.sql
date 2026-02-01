-- Migration: Add tenant_id to existing tables
-- This migration handles existing data by:
-- 1. Adding tenant_id as nullable
-- 2. Creating a default tenant for existing data
-- 3. Updating all records with the default tenant
-- 4. Making the column NOT NULL

-- Create a default tenant for existing data (if not exists)
DO $$
DECLARE
    default_tenant_id UUID;
    first_user_id UUID;
BEGIN
    -- Get the first user to be the owner of the default tenant
    SELECT id INTO first_user_id FROM users ORDER BY created_at LIMIT 1;
    
    -- If no users exist, skip the migration
    IF first_user_id IS NULL THEN
        RAISE NOTICE 'No users found, skipping tenant migration';
        RETURN;
    END IF;
    
    -- Check if default tenant already exists
    SELECT id INTO default_tenant_id FROM tenants WHERE slug = 'default-tenant' LIMIT 1;
    
    -- Create default tenant if it doesn't exist
    IF default_tenant_id IS NULL THEN
        default_tenant_id := gen_random_uuid();
        INSERT INTO tenants (id, name, slug, owner_id, settings, created_at, updated_at)
        VALUES (
            default_tenant_id,
            'Default Organization',
            'default-tenant',
            first_user_id,
            '{"currency": "KES", "timezone": "Africa/Nairobi", "feature_flags": {"inventory": true, "orders": true, "reports": true, "quotations": true}}'::jsonb,
            NOW(),
            NOW()
        );
        
        -- Add owner as member
        INSERT INTO tenant_memberships (tenant_id, user_id, role, created_at)
        VALUES (default_tenant_id, first_user_id, 'owner', NOW());
        
        RAISE NOTICE 'Created default tenant with ID: %', default_tenant_id;
    END IF;
    
    -- Add tenant_id column to products (if not exists)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'products' AND column_name = 'tenant_id') THEN
        ALTER TABLE products ADD COLUMN tenant_id UUID;
        UPDATE products SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
        ALTER TABLE products ALTER COLUMN tenant_id SET NOT NULL;
        CREATE INDEX IF NOT EXISTS idx_products_tenant_id ON products(tenant_id);
        RAISE NOTICE 'Added tenant_id to products';
    END IF;
    
    -- Add tenant_id column to categories (if not exists)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'categories' AND column_name = 'tenant_id') THEN
        ALTER TABLE categories ADD COLUMN tenant_id UUID;
        UPDATE categories SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
        ALTER TABLE categories ALTER COLUMN tenant_id SET NOT NULL;
        CREATE INDEX IF NOT EXISTS idx_categories_tenant_id ON categories(tenant_id);
        RAISE NOTICE 'Added tenant_id to categories';
    END IF;
    
    -- Add tenant_id column to units (if not exists)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'units' AND column_name = 'tenant_id') THEN
        ALTER TABLE units ADD COLUMN tenant_id UUID;
        UPDATE units SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
        ALTER TABLE units ALTER COLUMN tenant_id SET NOT NULL;
        CREATE INDEX IF NOT EXISTS idx_units_tenant_id ON units(tenant_id);
        RAISE NOTICE 'Added tenant_id to units';
    END IF;
    
    -- Add tenant_id column to orders (if not exists)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'orders' AND column_name = 'tenant_id') THEN
        ALTER TABLE orders ADD COLUMN tenant_id UUID;
        UPDATE orders SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
        ALTER TABLE orders ALTER COLUMN tenant_id SET NOT NULL;
        CREATE INDEX IF NOT EXISTS idx_orders_tenant_id ON orders(tenant_id);
        RAISE NOTICE 'Added tenant_id to orders';
    END IF;
    
    -- Add tenant_id column to customers (if not exists)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'customers' AND column_name = 'tenant_id') THEN
        ALTER TABLE customers ADD COLUMN tenant_id UUID;
        UPDATE customers SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
        ALTER TABLE customers ALTER COLUMN tenant_id SET NOT NULL;
        CREATE INDEX IF NOT EXISTS idx_customers_tenant_id ON customers(tenant_id);
        RAISE NOTICE 'Added tenant_id to customers';
    END IF;
    
    -- Add tenant_id column to suppliers (if not exists)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'suppliers' AND column_name = 'tenant_id') THEN
        ALTER TABLE suppliers ADD COLUMN tenant_id UUID;
        UPDATE suppliers SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
        ALTER TABLE suppliers ALTER COLUMN tenant_id SET NOT NULL;
        CREATE INDEX IF NOT EXISTS idx_suppliers_tenant_id ON suppliers(tenant_id);
        RAISE NOTICE 'Added tenant_id to suppliers';
    END IF;
    
    -- Add tenant_id column to purchases (if not exists)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'purchases' AND column_name = 'tenant_id') THEN
        ALTER TABLE purchases ADD COLUMN tenant_id UUID;
        UPDATE purchases SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
        ALTER TABLE purchases ALTER COLUMN tenant_id SET NOT NULL;
        CREATE INDEX IF NOT EXISTS idx_purchases_tenant_id ON purchases(tenant_id);
        RAISE NOTICE 'Added tenant_id to purchases';
    END IF;
    
    -- Add tenant_id column to quotations (if not exists)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'quotations' AND column_name = 'tenant_id') THEN
        ALTER TABLE quotations ADD COLUMN tenant_id UUID;
        UPDATE quotations SET tenant_id = default_tenant_id WHERE tenant_id IS NULL;
        ALTER TABLE quotations ALTER COLUMN tenant_id SET NOT NULL;
        CREATE INDEX IF NOT EXISTS idx_quotations_tenant_id ON quotations(tenant_id);
        RAISE NOTICE 'Added tenant_id to quotations';
    END IF;
    
    -- Add all existing users to the default tenant
    INSERT INTO tenant_memberships (tenant_id, user_id, role, created_at)
    SELECT default_tenant_id, users.id, 'member', NOW()
    FROM users
    WHERE users.id != first_user_id
    AND NOT EXISTS (
        SELECT 1 FROM tenant_memberships 
        WHERE tenant_id = default_tenant_id AND user_id = users.id
    );
    
    RAISE NOTICE 'Tenant migration completed successfully';
END $$;
