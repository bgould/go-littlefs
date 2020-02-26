#include "go_lfs.h"

struct lfs* go_lfs_new_lfs() {
    return malloc(sizeof(struct lfs));
}

struct lfs_config* go_lfs_new_lfs_config() {
    return malloc(sizeof(struct lfs_config));
}

struct lfs_config* go_lfs_set_callbacks(struct lfs_config *cfg) {
    cfg->read  = go_lfs_c_cb_read;
    cfg->prog  = go_lfs_c_cb_prog;
    cfg->erase = go_lfs_c_cb_erase;
    cfg->sync  = go_lfs_c_cb_sync;
    return cfg;
/*
    struct lfs_config* retval = malloc(sizeof(struct lfs_config));
    retval->context = cfg->context;
	retval->read  = go_lfs_c_cb_read;
	retval->prog  = go_lfs_c_cb_prog;
	retval->erase = go_lfs_c_cb_erase;
	retval->sync  = go_lfs_c_cb_sync;
	retval->read_size = cfg->read_size;
	retval->prog_size = cfg->prog_size;
	retval->block_size = cfg->block_size;
	retval->block_count = cfg->block_count;
	retval->cache_size = cfg->cache_size;
	retval->lookahead_size = cfg->lookahead_size;
	retval->block_cycles = cfg->block_cycles;
	return retval;
	*/
}

int go_lfs_c_cb_read(const struct lfs_config *c, lfs_block_t block, lfs_off_t off, void *buffer, lfs_size_t size) {
	return go_lfs_block_device_read(c->context, block, off, buffer, size);
}

int go_lfs_c_cb_prog(const struct lfs_config *c, lfs_block_t block, lfs_off_t off, const void *buffer, lfs_size_t size) {
	return go_lfs_block_device_prog(c->context, block, off, buffer, size);
}

int go_lfs_c_cb_erase(const struct lfs_config *c, lfs_block_t block) {
	return go_lfs_block_device_erase(c->context, block);
}

int go_lfs_c_cb_sync(const struct lfs_config *c) {
	return go_lfs_block_device_sync(c->context);
}