 -- 创建数据库
CREATE DATABASE blog
    ENCODING 'UTF8'
    LC_COLLATE 'zh_CN.UTF8'
    LC_CTYPE 'zh_CN.UTF8'
    TEMPLATE template0;

\c blog;

-- 安全策略
ALTER DEFAULT PRIVILEGES REVOKE EXECUTE ON FUNCTIONS FROM PUBLIC;
CREATE ROLE blog_test WITH LOGIN PASSWORD 'blogtestpass' NOINHERIT;
GRANT CONNECT ON DATABASE blog TO blog_test;
GRANT USAGE ON SCHEMA public TO blog_test;

-- 启用必要扩展
CREATE EXTENSION IF NOT EXISTS ltree;       -- 分类树结构支持
CREATE EXTENSION IF NOT EXISTS pgcrypto;    -- 密码加密
CREATE EXTENSION IF NOT EXISTS citext;      -- 不区分大小写文本
CREATE EXTENSION IF NOT EXISTS pg_trgm;     -- 文本搜索支持

-- 枚举定义表
CREATE TABLE IF NOT EXISTS sys_enums (
    enum_id SMALLSERIAL PRIMARY KEY,                      -- 枚举ID
    enum_type VARCHAR(20) NOT NULL,                       -- 枚举类型
    enum_value SMALLINT NOT NULL,                         -- 枚举值
    enum_name VARCHAR(50) NOT NULL,                       -- 枚举显示名称
    enum_desc VARCHAR(200),                               -- 枚举描述
    i18n JSONB NOT NULL DEFAULT '{}',                     -- 多语言支持
    sort_order SMALLINT NOT NULL DEFAULT 0,               -- 排序
    is_default BOOLEAN NOT NULL DEFAULT FALSE,            -- 是否默认值
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,             -- 是否启用
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 创建时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 更新时间
    UNIQUE (enum_type, enum_value)                        -- 类型+值唯一
);

COMMENT ON TABLE sys_enums IS '系统枚举定义表';
COMMENT ON COLUMN sys_enums.enum_id IS '枚举ID';
COMMENT ON COLUMN sys_enums.enum_type IS '枚举类型标识';
COMMENT ON COLUMN sys_enums.enum_value IS '枚举数值，用于存储';
COMMENT ON COLUMN sys_enums.enum_name IS '枚举名称，用于显示';
COMMENT ON COLUMN sys_enums.enum_desc IS '枚举详细描述';
COMMENT ON COLUMN sys_enums.i18n IS '多语言翻译，格式如：{"en":"Name","zh-TW":"名稱"}';
COMMENT ON COLUMN sys_enums.sort_order IS '显示排序值';
COMMENT ON COLUMN sys_enums.is_default IS '是否为默认选项';
COMMENT ON COLUMN sys_enums.is_enabled IS '是否启用该枚举值';
COMMENT ON COLUMN sys_enums.created_at IS '创建时间';
COMMENT ON COLUMN sys_enums.updated_at IS '更新时间';

-- 枚举表索引
CREATE INDEX idx_sys_enums_type ON sys_enums(enum_type);
CREATE INDEX idx_sys_enums_enabled ON sys_enums(is_enabled) WHERE is_enabled = TRUE;

-- 初始化枚举数据
INSERT INTO sys_enums (enum_type, enum_value, enum_name, enum_desc, is_default) VALUES
-- 用户状态枚举
('user_status', 1, '正常', '正常状态的用户', TRUE),
('user_status', 2, '禁用', '被管理员禁用的用户', FALSE),
('user_status', 3, '未激活', '注册但未激活的用户', FALSE),
-- 用户注册渠道枚举
('register_source', 1, '邮箱注册', '通过邮箱注册的用户', TRUE),
('register_source', 2, '手机注册', '通过手机号注册的用户', FALSE),
('register_source', 3, '微信登录', '通过微信授权登录的用户', FALSE),
-- 文章状态枚举
('article_status', 1, '草稿', '未完成的文章草稿', TRUE),
('article_status', 2, '待审核', '提交等待审核的文章', FALSE),
('article_status', 3, '已发布', '审核通过并发布的文章', FALSE),
('article_status', 4, '已下线', '被下线的文章', FALSE),
--
('comment_status', 1, '待审核', '评论待审核状态', TRUE),
('comment_status', 2, '已通过', '评论已通过状态', FALSE),
('comment_status', 3, '已拒绝', '评论已拒绝状态', FALSE),
-- 文章类型枚举
('article_type', 1, '原创', '原创内容', TRUE),
('article_type', 2, '转载', '转载的内容', FALSE),
('article_type', 3, '翻译', '翻译的内容', FALSE),
('article_type', 4, 'AI', 'AI生成的内容', FALSE);

-- 用户表
CREATE TABLE IF NOT EXISTS sys_users (
    user_id SERIAL PRIMARY KEY,                           -- 用户ID
    username VARCHAR(30) NOT NULL UNIQUE,                 -- 用户名
    password_hash VARCHAR(100),                           -- 密码哈希
    email VARCHAR(100) UNIQUE,                            -- 邮箱
    mobile VARCHAR(20) UNIQUE,                            -- 手机号
    wechat_openid VARCHAR(50) UNIQUE,                     -- 微信开放ID
    wechat_unionid VARCHAR(50) UNIQUE,                    -- 微信统一ID
    avatar VARCHAR(255),                                  -- 头像URL
    nickname VARCHAR(50),                                 -- 昵称
    real_name VARCHAR(50),                                -- 真实姓名
    gender SMALLINT DEFAULT 0,                            -- 性别(0未知,1男,2女)
    birthday DATE,                                        -- 生日
    status SMALLINT NOT NULL DEFAULT 1,                   -- 状态(1正常,2禁用,3未激活)
    register_source SMALLINT NOT NULL DEFAULT 1,          -- 注册来源(1邮箱,2手机,3微信)
    last_login TIMESTAMPTZ,                               -- 最后登录时间
    login_count INT NOT NULL DEFAULT 0,                   -- 登录次数
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 创建时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()         -- 更新时间
);

COMMENT ON TABLE sys_users IS '系统用户表';
COMMENT ON COLUMN sys_users.user_id IS '用户唯一标识';
COMMENT ON COLUMN sys_users.username IS '用户登录名，4-30位字母数字下划线组合';
COMMENT ON COLUMN sys_users.password_hash IS '使用pgcrypto加密的密码哈希';
COMMENT ON COLUMN sys_users.email IS '用户邮箱，可用于登录';
COMMENT ON COLUMN sys_users.mobile IS '用户手机号，可用于登录';
COMMENT ON COLUMN sys_users.wechat_openid IS '微信授权登录的OpenID';
COMMENT ON COLUMN sys_users.wechat_unionid IS '跨应用微信UnionID';
COMMENT ON COLUMN sys_users.avatar IS '用户头像图片地址';
COMMENT ON COLUMN sys_users.nickname IS '用户昵称，显示用';
COMMENT ON COLUMN sys_users.real_name IS '用户真实姓名';
COMMENT ON COLUMN sys_users.gender IS '性别：0未知，1男，2女';
COMMENT ON COLUMN sys_users.birthday IS '用户生日';
COMMENT ON COLUMN sys_users.status IS '用户状态：1正常，2禁用，3未激活';
COMMENT ON COLUMN sys_users.register_source IS '注册来源：1邮箱，2手机号，3微信';
COMMENT ON COLUMN sys_users.last_login IS '最后登录时间';
COMMENT ON COLUMN sys_users.login_count IS '累计登录次数';
COMMENT ON COLUMN sys_users.created_at IS '账号创建时间';
COMMENT ON COLUMN sys_users.updated_at IS '账号信息更新时间';

-- 用户表索引
CREATE INDEX idx_users_status ON sys_users(status);
CREATE INDEX idx_users_mobile ON sys_users(mobile) WHERE mobile IS NOT NULL;
CREATE INDEX idx_users_email ON sys_users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_login ON sys_users(last_login) WHERE last_login IS NOT NULL;
CREATE INDEX idx_users_wechat ON sys_users(wechat_openid) WHERE wechat_openid IS NOT NULL;

-- 登录日志表
CREATE TABLE IF NOT EXISTS sys_login_logs (
    log_id BIGSERIAL PRIMARY KEY,                         -- 日志ID
    user_id INT,                                         -- 用户ID
    username VARCHAR(30),                                -- 登录用户名
    login_type SMALLINT NOT NULL,                        -- 登录类型(1用户名,2邮箱,3手机,4微信)
    login_status SMALLINT NOT NULL,                      -- 登录结果(1成功,2失败)
    ip_address INET NOT NULL,                            -- 登录IP
    user_agent TEXT,                                      -- 用户代理
    device_info JSONB,                                   -- 设备信息
    login_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),       -- 登录时间
    fail_reason VARCHAR(100)                             -- 失败原因
) PARTITION BY RANGE (login_time);

COMMENT ON TABLE sys_login_logs IS '用户登录日志表';
COMMENT ON COLUMN sys_login_logs.log_id IS '日志唯一标识';
COMMENT ON COLUMN sys_login_logs.user_id IS '关联的用户ID';
COMMENT ON COLUMN sys_login_logs.username IS '登录时使用的用户名';
COMMENT ON COLUMN sys_login_logs.login_type IS '登录类型：1用户名密码，2邮箱，3手机号，4微信';
COMMENT ON COLUMN sys_login_logs.login_status IS '登录状态：1成功，2失败';
COMMENT ON COLUMN sys_login_logs.ip_address IS '登录IP地址';
COMMENT ON COLUMN sys_login_logs.user_agent IS '浏览器代理信息';
COMMENT ON COLUMN sys_login_logs.device_info IS '登录设备信息JSON';
COMMENT ON COLUMN sys_login_logs.login_time IS '登录时间';
COMMENT ON COLUMN sys_login_logs.fail_reason IS '登录失败原因';

-- 登录日志默认分区
CREATE TABLE sys_login_logs_2024 PARTITION OF sys_login_logs
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

-- 登录日志索引
CREATE INDEX idx_login_logs_user ON sys_login_logs(user_id);
CREATE INDEX idx_login_logs_time ON sys_login_logs USING BRIN(login_time);
CREATE INDEX idx_login_logs_ip ON sys_login_logs(ip_address);
CREATE INDEX idx_login_logs_status ON sys_login_logs(login_status);

-- 角色表
CREATE TABLE IF NOT EXISTS sys_roles (
    role_id SERIAL PRIMARY KEY,                           -- 角色ID
    role_name VARCHAR(50) NOT NULL UNIQUE,                -- 角色名称
    role_key VARCHAR(50) NOT NULL UNIQUE,                 -- 角色标识
    role_sort SMALLINT NOT NULL DEFAULT 0,                -- 角色排序
    role_desc VARCHAR(200),                               -- 角色描述
    is_default BOOLEAN NOT NULL DEFAULT FALSE,            -- 是否默认角色
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,             -- 是否启用
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 创建时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()         -- 更新时间
);

COMMENT ON TABLE sys_roles IS '系统角色表';
COMMENT ON COLUMN sys_roles.role_id IS '角色唯一标识';
COMMENT ON COLUMN sys_roles.role_name IS '角色名称，如管理员、编辑等';
COMMENT ON COLUMN sys_roles.role_key IS '角色唯一标识符，用于程序中识别角色';
COMMENT ON COLUMN sys_roles.role_sort IS '角色排序字段';
COMMENT ON COLUMN sys_roles.role_desc IS '角色详细描述';
COMMENT ON COLUMN sys_roles.is_default IS '是否为新用户默认角色';
COMMENT ON COLUMN sys_roles.is_enabled IS '角色是否启用';
COMMENT ON COLUMN sys_roles.created_at IS '角色创建时间';
COMMENT ON COLUMN sys_roles.updated_at IS '角色更新时间';

-- 角色表索引
CREATE INDEX idx_roles_enabled ON sys_roles(is_enabled);
CREATE INDEX idx_roles_sort ON sys_roles(role_sort);

-- 初始化角色数据
INSERT INTO sys_roles (role_name, role_key, role_sort, role_desc, is_default, is_enabled) VALUES
('超级管理员', 'admin', 1, '系统超级管理员，拥有所有权限', FALSE, TRUE),
('普通用户', 'user', 5, '普通注册用户，拥有基本权限', TRUE, TRUE),
('内容编辑', 'editor', 3, '内容编辑人员，负责内容管理', FALSE, TRUE),
('审核员', 'reviewer', 2, '内容审核人员，负责审核内容', FALSE, TRUE);

-- 权限表
CREATE TABLE IF NOT EXISTS sys_permissions (
    perm_id SERIAL PRIMARY KEY,                           -- 权限ID
    perm_name VARCHAR(50) NOT NULL,                       -- 权限名称
    perm_key VARCHAR(50) NOT NULL UNIQUE,                 -- 权限标识
    perm_type SMALLINT NOT NULL DEFAULT 1,                -- 权限类型(1菜单,2按钮,3接口)
    parent_id INT,                                        -- 父权限ID
    path LTREE,                                           -- 权限路径
    api_path VARCHAR(200),                                -- API路径
    component VARCHAR(100),                               -- 前端组件
    perms VARCHAR(100),                                   -- 权限字符串
    icon VARCHAR(100),                                    -- 图标
    menu_sort SMALLINT NOT NULL DEFAULT 0,                -- 菜单排序
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,             -- 是否可见
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,             -- 是否启用
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 创建时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()         -- 更新时间
);

COMMENT ON TABLE sys_permissions IS '系统权限表';
COMMENT ON COLUMN sys_permissions.perm_id IS '权限唯一标识';
COMMENT ON COLUMN sys_permissions.perm_name IS '权限名称，显示用';
COMMENT ON COLUMN sys_permissions.perm_key IS '权限唯一标识符，用于程序中识别权限';
COMMENT ON COLUMN sys_permissions.perm_type IS '权限类型：1菜单，2按钮，3接口';
COMMENT ON COLUMN sys_permissions.parent_id IS '父权限ID，构建权限树';
COMMENT ON COLUMN sys_permissions.path IS '使用ltree存储的权限路径';
COMMENT ON COLUMN sys_permissions.api_path IS 'API接口路径';
COMMENT ON COLUMN sys_permissions.component IS '前端组件路径';
COMMENT ON COLUMN sys_permissions.perms IS '权限标识字符串';
COMMENT ON COLUMN sys_permissions.icon IS '图标样式或地址';
COMMENT ON COLUMN sys_permissions.menu_sort IS '菜单显示排序';
COMMENT ON COLUMN sys_permissions.is_visible IS '是否在菜单中显示';
COMMENT ON COLUMN sys_permissions.is_enabled IS '权限是否启用';
COMMENT ON COLUMN sys_permissions.created_at IS '权限创建时间';
COMMENT ON COLUMN sys_permissions.updated_at IS '权限更新时间';

-- 权限表索引
CREATE INDEX idx_permissions_parent ON sys_permissions(parent_id);
CREATE INDEX idx_permissions_type ON sys_permissions(perm_type);
CREATE INDEX idx_permissions_enabled ON sys_permissions(is_enabled);
CREATE INDEX idx_permissions_visible ON sys_permissions(is_visible);
CREATE INDEX idx_permissions_path_gist ON sys_permissions USING GIST (path);

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS sys_user_roles (
    user_id INT NOT NULL,                                -- 用户ID
    role_id INT NOT NULL,                                -- 角色ID
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),       -- 创建时间
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES sys_users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES sys_roles(role_id) ON DELETE CASCADE
);

COMMENT ON TABLE sys_user_roles IS '用户角色关联表';
COMMENT ON COLUMN sys_user_roles.user_id IS '关联的用户ID';
COMMENT ON COLUMN sys_user_roles.role_id IS '关联的角色ID';
COMMENT ON COLUMN sys_user_roles.created_at IS '关联创建时间';

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS sys_role_permissions (
    role_id INT NOT NULL,                                -- 角色ID
    perm_id INT NOT NULL,                                -- 权限ID
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),       -- 创建时间
    PRIMARY KEY (role_id, perm_id),
    FOREIGN KEY (role_id) REFERENCES sys_roles(role_id) ON DELETE CASCADE,
    FOREIGN KEY (perm_id) REFERENCES sys_permissions(perm_id) ON DELETE CASCADE
);

COMMENT ON TABLE sys_role_permissions IS '角色权限关联表';
COMMENT ON COLUMN sys_role_permissions.role_id IS '关联的角色ID';
COMMENT ON COLUMN sys_role_permissions.perm_id IS '关联的权限ID';
COMMENT ON COLUMN sys_role_permissions.created_at IS '关联创建时间';

-- 分类表
CREATE TABLE IF NOT EXISTS cms_categories (
    category_id SERIAL PRIMARY KEY,                       -- 分类ID
    parent_id INT,                                        -- 父分类ID
    category_name VARCHAR(50) NOT NULL,                   -- 分类名称
    category_key VARCHAR(50) NOT NULL UNIQUE,             -- 分类标识
    path LTREE NOT NULL,                                  -- 分类路径
    description TEXT,                                     -- 分类描述
    thumbnail VARCHAR(255),                               -- 分类缩略图
    icon VARCHAR(100),                                    -- 分类图标
    sort_order SMALLINT NOT NULL DEFAULT 0,               -- 排序
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,             -- 是否显示
    seo_title VARCHAR(100),                               -- SEO标题
    seo_keywords VARCHAR(200),                            -- SEO关键词
    seo_description VARCHAR(300),                         -- SEO描述
    article_count INT NOT NULL DEFAULT 0,                 -- 文章数量
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 创建时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 更新时间
    FOREIGN KEY (parent_id) REFERENCES cms_categories(category_id) ON DELETE CASCADE
);

COMMENT ON TABLE cms_categories IS '文章分类表';
COMMENT ON COLUMN cms_categories.category_id IS '分类唯一标识';
COMMENT ON COLUMN cms_categories.parent_id IS '父分类ID，构建分类树';
COMMENT ON COLUMN cms_categories.category_name IS '分类名称';
COMMENT ON COLUMN cms_categories.category_key IS '分类标识符，URL友好的格式';
COMMENT ON COLUMN cms_categories.path IS '使用ltree存储的分类路径';
COMMENT ON COLUMN cms_categories.description IS '分类描述';
COMMENT ON COLUMN cms_categories.thumbnail IS '分类缩略图URL';
COMMENT ON COLUMN cms_categories.icon IS '分类图标';
COMMENT ON COLUMN cms_categories.sort_order IS '分类排序字段';
COMMENT ON COLUMN cms_categories.is_visible IS '分类是否在前台可见';
COMMENT ON COLUMN cms_categories.seo_title IS 'SEO优化标题';
COMMENT ON COLUMN cms_categories.seo_keywords IS 'SEO优化关键词';
COMMENT ON COLUMN cms_categories.seo_description IS 'SEO优化描述';
COMMENT ON COLUMN cms_categories.article_count IS '该分类下的文章数量';
COMMENT ON COLUMN cms_categories.created_at IS '分类创建时间';
COMMENT ON COLUMN cms_categories.updated_at IS '分类更新时间';

-- 分类表索引
CREATE INDEX idx_categories_parent ON cms_categories(parent_id);
CREATE INDEX idx_categories_visible ON cms_categories(is_visible);
CREATE INDEX idx_categories_path_gist ON cms_categories USING GIST (path);
CREATE INDEX idx_categories_sort ON cms_categories(sort_order);

-- 标签表
CREATE TABLE IF NOT EXISTS cms_tags (
    tag_id SERIAL PRIMARY KEY,                            -- 标签ID
    tag_name VARCHAR(50) NOT NULL UNIQUE,                 -- 标签名称
    tag_key VARCHAR(50) NOT NULL UNIQUE,                  -- 标签标识
    description TEXT,                                     -- 标签描述
    thumbnail VARCHAR(255),                               -- 标签缩略图
    sort_order SMALLINT NOT NULL DEFAULT 0,               -- 排序
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,             -- 是否显示
    article_count INT NOT NULL DEFAULT 0,                 -- 文章数量
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 创建时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()         -- 更新时间
);

COMMENT ON TABLE cms_tags IS '文章标签表';
COMMENT ON COLUMN cms_tags.tag_id IS '标签唯一标识';
COMMENT ON COLUMN cms_tags.tag_name IS '标签名称';
COMMENT ON COLUMN cms_tags.tag_key IS '标签标识符，URL友好的格式';
COMMENT ON COLUMN cms_tags.description IS '标签描述';
COMMENT ON COLUMN cms_tags.thumbnail IS '标签缩略图URL';
COMMENT ON COLUMN cms_tags.sort_order IS '标签排序字段';
COMMENT ON COLUMN cms_tags.is_visible IS '标签是否在前台可见';
COMMENT ON COLUMN cms_tags.article_count IS '使用该标签的文章数量';
COMMENT ON COLUMN cms_tags.created_at IS '标签创建时间';
COMMENT ON COLUMN cms_tags.updated_at IS '标签更新时间';

-- 标签表索引
CREATE INDEX idx_tags_visible ON cms_tags(is_visible);
CREATE INDEX idx_tags_count ON cms_tags(article_count);
CREATE INDEX idx_tags_sort ON cms_tags(sort_order);

-- 文章表
CREATE TABLE IF NOT EXISTS cms_articles (
    article_id BIGSERIAL PRIMARY KEY,                     -- 文章ID
    user_id INT NOT NULL,                                 -- 作者ID
    title VARCHAR(200) NOT NULL,                          -- 文章标题
    article_key VARCHAR(200) NOT NULL UNIQUE,             -- 文章标识
    summary VARCHAR(500),                                 -- 文章摘要
    thumbnail VARCHAR(255),                               -- 文章缩略图
    status SMALLINT NOT NULL DEFAULT 1,                   -- 文章状态(1草稿,2待审核,3已发布,4已下线)
    article_type SMALLINT NOT NULL DEFAULT 1,             -- 文章类型(1原创,2转载,3翻译)
    view_count INT NOT NULL DEFAULT 0,                    -- 查看次数
    like_count INT NOT NULL DEFAULT 0,                    -- 点赞次数
    comment_count INT NOT NULL DEFAULT 0,                 -- 评论次数
    allow_comment BOOLEAN NOT NULL DEFAULT TRUE,          -- 是否允许评论
    is_top BOOLEAN NOT NULL DEFAULT FALSE,                -- 是否置顶
    is_recommend BOOLEAN NOT NULL DEFAULT FALSE,          -- 是否推荐
    seo_title VARCHAR(100),                               -- SEO标题
    seo_keywords VARCHAR(200),                            -- SEO关键词
    seo_description VARCHAR(300),                         -- SEO描述
    source_url VARCHAR(255),                              -- 原文链接(转载/翻译)
    source_name VARCHAR(100),                             -- 来源名称
    publish_time TIMESTAMPTZ,                             -- 发布时间
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 创建时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 更新时间
    FOREIGN KEY (user_id) REFERENCES sys_users(user_id)
) PARTITION BY RANGE (created_at);

COMMENT ON TABLE cms_articles IS '文章主表';
COMMENT ON COLUMN cms_articles.article_id IS '文章唯一标识';
COMMENT ON COLUMN cms_articles.user_id IS '文章作者ID';
COMMENT ON COLUMN cms_articles.title IS '文章标题';
COMMENT ON COLUMN cms_articles.article_key IS '文章唯一标识符，URL友好的格式';
COMMENT ON COLUMN cms_articles.summary IS '文章摘要';
COMMENT ON COLUMN cms_articles.thumbnail IS '文章缩略图URL';
COMMENT ON COLUMN cms_articles.status IS '文章状态：1草稿，2待审核，3已发布，4已下线';
COMMENT ON COLUMN cms_articles.article_type IS '文章类型：1原创，2转载，3翻译';
COMMENT ON COLUMN cms_articles.view_count IS '文章查看次数';
COMMENT ON COLUMN cms_articles.like_count IS '文章点赞次数';
COMMENT ON COLUMN cms_articles.comment_count IS '文章评论次数';
COMMENT ON COLUMN cms_articles.allow_comment IS '是否允许评论';
COMMENT ON COLUMN cms_articles.is_top IS '是否置顶显示';
COMMENT ON COLUMN cms_articles.is_recommend IS '是否推荐';
COMMENT ON COLUMN cms_articles.seo_title IS 'SEO优化标题';
COMMENT ON COLUMN cms_articles.seo_keywords IS 'SEO优化关键词';
COMMENT ON COLUMN cms_articles.seo_description IS 'SEO优化描述';
COMMENT ON COLUMN cms_articles.source_url IS '转载或翻译的原文链接';
COMMENT ON COLUMN cms_articles.source_name IS '转载或翻译的来源名称';
COMMENT ON COLUMN cms_articles.publish_time IS '文章发布时间';
COMMENT ON COLUMN cms_articles.created_at IS '文章创建时间';
COMMENT ON COLUMN cms_articles.updated_at IS '文章更新时间';

-- 文章默认分区
CREATE TABLE cms_articles_2024 PARTITION OF cms_articles
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

-- 文章表索引
CREATE INDEX idx_articles_user ON cms_articles(user_id);
CREATE INDEX idx_articles_status ON cms_articles(status);
CREATE INDEX idx_articles_type ON cms_articles(article_type);
CREATE INDEX idx_articles_created ON cms_articles USING BRIN(created_at);
CREATE INDEX idx_articles_published ON cms_articles(publish_time) WHERE status = 3;
CREATE INDEX idx_articles_recommend ON cms_articles(is_recommend) WHERE is_recommend = TRUE;
CREATE INDEX idx_articles_top ON cms_articles(is_top) WHERE is_top = TRUE;

-- 文章内容表
CREATE TABLE IF NOT EXISTS cms_article_contents (
    content_id BIGSERIAL PRIMARY KEY,                     -- 内容ID
    article_id BIGINT NOT NULL,                           -- 文章ID
    content TEXT NOT NULL,                                -- 文章内容
    content_format SMALLINT NOT NULL DEFAULT 1,           -- 内容格式(1Markdown,2HTML)
    version INT NOT NULL DEFAULT 1,                       -- 内容版本
    is_current BOOLEAN NOT NULL DEFAULT TRUE,             -- 是否当前版本
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 创建时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 更新时间
    FOREIGN KEY (article_id) REFERENCES cms_articles(article_id) ON DELETE CASCADE
) PARTITION BY RANGE (created_at);

COMMENT ON TABLE cms_article_contents IS '文章内容表';
COMMENT ON COLUMN cms_article_contents.content_id IS '内容记录唯一标识';
COMMENT ON COLUMN cms_article_contents.article_id IS '关联的文章ID';
COMMENT ON COLUMN cms_article_contents.content IS '文章完整内容';
COMMENT ON COLUMN cms_article_contents.content_format IS '内容格式：1Markdown，2HTML';
COMMENT ON COLUMN cms_article_contents.version IS '内容版本号';
COMMENT ON COLUMN cms_article_contents.is_current IS '是否为当前版本';
COMMENT ON COLUMN cms_article_contents.created_at IS '内容创建时间';
COMMENT ON COLUMN cms_article_contents.updated_at IS '内容更新时间';

-- 文章内容默认分区
CREATE TABLE IF NOT EXISTS cms_article_contents_current PARTITION OF cms_article_contents
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

-- 文章内容表索引
CREATE INDEX idx_article_contents_article ON cms_article_contents(article_id);
CREATE INDEX idx_article_contents_current ON cms_article_contents(is_current) WHERE is_current = TRUE;
CREATE INDEX idx_article_contents_time ON cms_article_contents USING BRIN(created_at);

-- 文章-分类关联表
CREATE TABLE IF NOT EXISTS cms_article_categories (
    article_id BIGINT NOT NULL,                           -- 文章ID
    category_id INT NOT NULL,                             -- 分类ID
    is_primary BOOLEAN NOT NULL DEFAULT FALSE,            -- 是否主分类
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 创建时间
    PRIMARY KEY (article_id, category_id),
    FOREIGN KEY (article_id) REFERENCES cms_articles(article_id) ON DELETE CASCADE,
    FOREIGN KEY (category_id) REFERENCES cms_categories(category_id) ON DELETE CASCADE
);

COMMENT ON TABLE cms_article_categories IS '文章分类关联表';
COMMENT ON COLUMN cms_article_categories.article_id IS '关联的文章ID';
COMMENT ON COLUMN cms_article_categories.category_id IS '关联的分类ID';
COMMENT ON COLUMN cms_article_categories.is_primary IS '是否为文章的主分类';
COMMENT ON COLUMN cms_article_categories.created_at IS '关联创建时间';

-- 文章-标签关联表
CREATE TABLE IF NOT EXISTS cms_article_tags (
    article_id BIGINT NOT NULL,                           -- 文章ID
    tag_id INT NOT NULL,                                  -- 标签ID
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 创建时间
    PRIMARY KEY (article_id, tag_id),
    FOREIGN KEY (article_id) REFERENCES cms_articles(article_id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES cms_tags(tag_id) ON DELETE CASCADE
);

COMMENT ON TABLE cms_article_tags IS '文章标签关联表';
COMMENT ON COLUMN cms_article_tags.article_id IS '关联的文章ID';
COMMENT ON COLUMN cms_article_tags.tag_id IS '关联的标签ID';
COMMENT ON COLUMN cms_article_tags.created_at IS '关联创建时间';

-- 系统配置表
CREATE TABLE IF NOT EXISTS sys_configs (
    config_id SERIAL PRIMARY KEY,                         -- 配置ID
    config_name VARCHAR(100) NOT NULL,                    -- 配置名称
    config_key VARCHAR(100) NOT NULL UNIQUE,              -- 配置键
    config_value TEXT NOT NULL,                           -- 配置值
    value_type SMALLINT NOT NULL DEFAULT 1,               -- 值类型(1字符串,2数字,3布尔,4JSON)
    config_group VARCHAR(50) NOT NULL DEFAULT 'default',  -- 配置分组
    is_builtin BOOLEAN NOT NULL DEFAULT FALSE,            -- 是否内置
    is_frontend BOOLEAN NOT NULL DEFAULT FALSE,           -- 是否前端可用
    sort_order SMALLINT NOT NULL DEFAULT 0,               -- 排序
    remark VARCHAR(200),                                  -- 备注
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),        -- 创建时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()         -- 更新时间
);

COMMENT ON TABLE sys_configs IS '系统配置表';
COMMENT ON COLUMN sys_configs.config_id IS '配置唯一标识';
COMMENT ON COLUMN sys_configs.config_name IS '配置项名称';
COMMENT ON COLUMN sys_configs.config_key IS '配置键名，唯一标识符';
COMMENT ON COLUMN sys_configs.config_value IS '配置项值';
COMMENT ON COLUMN sys_configs.value_type IS '值类型：1字符串，2数字，3布尔，4JSON';
COMMENT ON COLUMN sys_configs.config_group IS '配置分组，便于管理';
COMMENT ON COLUMN sys_configs.is_builtin IS '是否为系统内置项，内置项不可删除';
COMMENT ON COLUMN sys_configs.is_frontend IS '是否允许前端访问';
COMMENT ON COLUMN sys_configs.sort_order IS '显示排序';
COMMENT ON COLUMN sys_configs.remark IS '配置说明';
COMMENT ON COLUMN sys_configs.created_at IS '配置创建时间';
COMMENT ON COLUMN sys_configs.updated_at IS '配置更新时间';

-- 系统配置表索引
CREATE INDEX idx_configs_group ON sys_configs(config_group);
CREATE INDEX idx_configs_frontend ON sys_configs(is_frontend) WHERE is_frontend = TRUE;
CREATE INDEX idx_configs_sort ON sys_configs(sort_order);

-- 初始化系统配置
INSERT INTO sys_configs (config_name, config_key, config_value, value_type, config_group, is_builtin, is_frontend, remark) VALUES
('网站名称', 'site_name', '我的博客', 1, 'site', TRUE, TRUE, '网站名称，显示在浏览器标题'),
('网站描述', 'site_description', '这是一个基于PostgreSQL的博客系统', 1, 'site', TRUE, TRUE, '网站描述信息'),
('网站关键词', 'site_keywords', '博客,PostgreSQL,技术博客', 1, 'site', TRUE, TRUE, '网站SEO关键词'),
('网站Logo', 'site_logo', '/static/images/logo.png', 1, 'site', TRUE, TRUE, '网站Logo图片路径'),
('ICP备案号', 'site_icp', '', 1, 'site', TRUE, TRUE, '网站ICP备案号'),
('公安备案号', 'site_police', '', 1, 'site', TRUE, TRUE, '网站公安备案号'),
('版权信息', 'site_copyright', '© 2024 My Blog', 1, 'site', TRUE, TRUE, '网站版权信息'),
('每页文章数', 'article_page_size', '10', 2, 'article', TRUE, TRUE, '文章列表每页显示数量');

-- 操作日志表
CREATE TABLE IF NOT EXISTS sys_operation_logs (
    log_id BIGSERIAL PRIMARY KEY,                         -- 日志ID
    user_id INT,                                          -- 用户ID
    username VARCHAR(30),                                 -- 操作用户名
    operation_type VARCHAR(50) NOT NULL,                  -- 操作类型
    operation_module VARCHAR(50) NOT NULL,                -- 操作模块
    operation_desc VARCHAR(200),                          -- 操作描述
    request_method VARCHAR(10),                           -- 请求方法
    request_url VARCHAR(255),                             -- 请求URL
    request_params TEXT,                                  -- 请求参数
    operation_result VARCHAR(10) NOT NULL, -- 操作结果(成功/失败)
    ip_address INET NOT NULL, -- 操作IP
    user_agent TEXT, -- 用户代理
    execution_time INT, -- 执行时间(毫秒)
    error_message TEXT, -- 错误信息
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW() -- 操作时间
    ) PARTITION BY RANGE (created_at);

COMMENT ON TABLE sys_operation_logs IS '系统操作日志表';
COMMENT ON COLUMN sys_operation_logs.log_id IS '日志唯一标识';
COMMENT ON COLUMN sys_operation_logs.user_id IS '操作用户ID';
COMMENT ON COLUMN sys_operation_logs.username IS '操作用户名';
COMMENT ON COLUMN sys_operation_logs.operation_type IS '操作类型，如：新增、修改、删除等';
COMMENT ON COLUMN sys_operation_logs.operation_module IS '操作模块，如：用户管理、文章管理等';
COMMENT ON COLUMN sys_operation_logs.operation_desc IS '操作描述';
COMMENT ON COLUMN sys_operation_logs.request_method IS 'HTTP请求方法，如：GET、POST等';
COMMENT ON COLUMN sys_operation_logs.request_url IS '请求URL地址';
COMMENT ON COLUMN sys_operation_logs.request_params IS '请求参数内容';
COMMENT ON COLUMN sys_operation_logs.operation_result IS '操作结果：success、fail';
COMMENT ON COLUMN sys_operation_logs.ip_address IS '操作者IP地址';
COMMENT ON COLUMN sys_operation_logs.user_agent IS '操作者浏览器代理信息';
COMMENT ON COLUMN sys_operation_logs.execution_time IS '操作执行时间（毫秒）';
COMMENT ON COLUMN sys_operation_logs.error_message IS '操作失败时的错误信息';
COMMENT ON COLUMN sys_operation_logs.created_at IS '操作发生时间';
-- 操作日志默认分区
CREATE TABLE sys_operation_logs_2024 PARTITION OF sys_operation_logs
FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');
-- 操作日志索引
CREATE INDEX idx_operation_logs_user ON sys_operation_logs(user_id);
CREATE INDEX idx_operation_logs_type ON sys_operation_logs(operation_type, operation_module);
CREATE INDEX idx_operation_logs_result ON sys_operation_logs(operation_result);
CREATE INDEX idx_operation_logs_time ON sys_operation_logs USING BRIN(created_at);
CREATE INDEX idx_operation_logs_ip ON sys_operation_logs(ip_address);
-- 评论表
CREATE TABLE IF NOT EXISTS cms_comments (
    comment_id BIGSERIAL PRIMARY KEY, -- 评论ID
    article_id BIGINT NOT NULL, -- 文章ID
    user_id INT NOT NULL, -- 评论用户ID
    parent_id BIGINT, -- 父评论ID
    root_id BIGINT, -- 根评论ID
    content TEXT NOT NULL, -- 评论内容
    ip_address INET, -- 评论IP
    user_agent TEXT, -- 用户代理
    liked_count INT NOT NULL DEFAULT 0, -- 点赞数
    is_approved BOOLEAN NOT NULL DEFAULT FALSE, -- 是否审核通过
    is_admin_reply BOOLEAN NOT NULL DEFAULT FALSE, -- 是否管理员回复
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 创建时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 更新时间
    FOREIGN KEY (article_id) REFERENCES cms_articles(article_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES sys_users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES cms_comments(comment_id) ON DELETE CASCADE
    );

COMMENT ON TABLE cms_comments IS '文章评论表';
COMMENT ON COLUMN cms_comments.comment_id IS '评论唯一标识';
COMMENT ON COLUMN cms_comments.article_id IS '关联的文章ID';
COMMENT ON COLUMN cms_comments.user_id IS '评论用户ID，未登录用户为NULL';
COMMENT ON COLUMN cms_comments.parent_id IS '回复的父评论ID';
COMMENT ON COLUMN cms_comments.root_id IS '评论树的根评论ID';
COMMENT ON COLUMN cms_comments.content IS '评论内容';
COMMENT ON COLUMN cms_comments.ip_address IS '评论者IP地址';
COMMENT ON COLUMN cms_comments.user_agent IS '评论者浏览器代理信息';
COMMENT ON COLUMN cms_comments.liked_count IS '评论获赞数量';
COMMENT ON COLUMN cms_comments.is_approved IS '评论是否已审核通过';
COMMENT ON COLUMN cms_comments.is_admin_reply IS '是否管理员回复';
COMMENT ON COLUMN cms_comments.created_at IS '评论创建时间';
COMMENT ON COLUMN cms_comments.updated_at IS '评论更新时间';


-- 评论表索引
CREATE INDEX idx_comments_article ON cms_comments(article_id);
CREATE INDEX idx_comments_user ON cms_comments(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_comments_parent ON cms_comments(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_comments_root ON cms_comments(root_id) WHERE root_id IS NOT NULL;
CREATE INDEX idx_comments_approved ON cms_comments(is_approved) WHERE is_approved = TRUE;
CREATE INDEX idx_comments_time ON cms_comments(created_at);

-- 文件存储表
CREATE TABLE IF NOT EXISTS sys_files (
    file_id BIGSERIAL PRIMARY KEY, -- 文件ID
    user_id INT, -- 上传用户ID
    original_name VARCHAR(255) NOT NULL, -- 原始文件名
    file_name VARCHAR(100) NOT NULL, -- 存储文件名
    file_path VARCHAR(255) NOT NULL, -- 文件路径
    file_ext VARCHAR(20) NOT NULL, -- 文件扩展名
    file_size BIGINT NOT NULL, -- 文件大小(字节)
    mime_type VARCHAR(100) NOT NULL, -- MIME类型
    storage_type SMALLINT NOT NULL DEFAULT 1, -- 存储类型(1本地,2云存储)
    use_times INT NOT NULL DEFAULT 0, -- 使用次数
    is_public BOOLEAN NOT NULL DEFAULT TRUE, -- 是否公开访问
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 创建时间
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- 更新时间
    FOREIGN KEY (user_id) REFERENCES sys_users(user_id) ON DELETE SET NULL
    );
COMMENT ON TABLE sys_files IS '文件存储表';
COMMENT ON COLUMN sys_files.file_id IS '文件唯一标识';
COMMENT ON COLUMN sys_files.user_id IS '上传文件的用户ID';
COMMENT ON COLUMN sys_files.original_name IS '文件原始名称';
COMMENT ON COLUMN sys_files.file_name IS '存储的文件名';
COMMENT ON COLUMN sys_files.file_path IS '文件存储路径';
COMMENT ON COLUMN sys_files.file_ext IS '文件扩展名，不含点';
COMMENT ON COLUMN sys_files.file_size IS '文件大小（字节）';
COMMENT ON COLUMN sys_files.mime_type IS '文件MIME类型';
COMMENT ON COLUMN sys_files.storage_type IS '存储类型：1本地存储，2云存储';
COMMENT ON COLUMN sys_files.use_times IS '文件引用次数';
COMMENT ON COLUMN sys_files.is_public IS '是否允许公开访问';
COMMENT ON COLUMN sys_files.created_at IS '上传时间';
COMMENT ON COLUMN sys_files.updated_at IS '更新时间';

-- 文件表索引
CREATE INDEX idx_files_user ON sys_files(user_id);
CREATE INDEX idx_files_ext ON sys_files(file_ext);
CREATE INDEX idx_files_time ON sys_files(created_at);
CREATE INDEX idx_files_public ON sys_files(is_public) WHERE is_public = TRUE;

-- 触发器函数定义
-- 更新分类路径触发器函数
CREATE OR REPLACE FUNCTION update_category_path() RETURNS TRIGGER AS $$
BEGIN
IF NEW.parent_id IS NULL THEN
-- 根分类
NEW.path = ltree(NEW.category_id::TEXT);
ELSE
-- 检查父分类是否存在
PERFORM 1 FROM cms_categories WHERE category_id = NEW.parent_id;
IF NOT FOUND THEN
RAISE EXCEPTION '父分类ID % 不存在', NEW.parent_id;
END IF;
-- 设置分类路径
SELECT path || NEW.category_id::TEXT INTO NEW.path
FROM cms_categories
WHERE category_id = NEW.parent_id;
END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 更新分类文章计数触发器函数
CREATE OR REPLACE FUNCTION update_category_article_count() RETURNS TRIGGER AS $$
BEGIN
IF TG_OP = 'INSERT' THEN
UPDATE cms_categories
SET article_count = article_count + 1
WHERE category_id = NEW.category_id;
ELSIF TG_OP = 'DELETE' THEN
UPDATE cms_categories
SET article_count = GREATEST(0, article_count - 1)
WHERE category_id = OLD.category_id;
END IF;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 更新标签文章计数触发器函数
CREATE OR REPLACE FUNCTION update_tag_article_count() RETURNS TRIGGER AS $$
BEGIN
IF TG_OP = 'INSERT' THEN
UPDATE cms_tags
SET article_count = article_count + 1
WHERE tag_id = NEW.tag_id;
ELSIF TG_OP = 'DELETE' THEN
UPDATE cms_tags
SET article_count = GREATEST(0, article_count - 1)
WHERE tag_id = OLD.tag_id;
END IF;
RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- 文章发布状态变更触发器函数
CREATE OR REPLACE FUNCTION article_status_change() RETURNS TRIGGER AS $$
BEGIN
-- 如果文章状态变更为已发布，且之前不是已发布状态
IF NEW.status = 3 AND (OLD.status IS NULL OR OLD.status != 3) THEN
-- 设置发布时间
NEW.publish_time = NOW();
END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 自动更新时间戳触发器函数
CREATE OR REPLACE FUNCTION update_timestamp() RETURNS TRIGGER AS $$
BEGIN
NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 触发器定义（所有触发器放在文件最后）
-- 分类路径触发器
CREATE TRIGGER trg_category_path
BEFORE INSERT OR UPDATE OF parent_id ON cms_categories
FOR EACH ROW EXECUTE FUNCTION update_category_path();

-- 分类文章计数触发器
CREATE TRIGGER trg_category_article_count
AFTER INSERT OR DELETE ON cms_article_categories
FOR EACH ROW EXECUTE FUNCTION update_category_article_count();

-- 标签文章计数触发器
CREATE TRIGGER trg_tag_article_count
AFTER INSERT OR DELETE ON cms_article_tags
FOR EACH ROW EXECUTE FUNCTION update_tag_article_count();

-- 文章状态变更触发器
CREATE TRIGGER trg_article_status_change
BEFORE UPDATE OF status ON cms_articles
FOR EACH ROW EXECUTE FUNCTION article_status_change();

-- 自动更新时间戳触发器
CREATE TRIGGER trg_users_timestamp
BEFORE UPDATE ON sys_users
FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER trg_articles_timestamp
BEFORE UPDATE ON cms_articles
FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER trg_categories_timestamp
BEFORE UPDATE ON cms_categories
FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER trg_tags_timestamp
BEFORE UPDATE ON cms_tags
FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER trg_comments_timestamp
BEFORE UPDATE ON cms_comments
FOR EACH ROW EXECUTE FUNCTION update_timestamp();
CREATE TRIGGER trg_configs_timestamp
BEFORE UPDATE ON sys_configs
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- 权限授予
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO blog_test;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO blog_test;