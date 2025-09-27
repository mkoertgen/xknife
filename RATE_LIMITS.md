# X-Knife Rate Limit Survival Guide

## The Twitter API Rate Limit Reality

Twitter API v2 has EXTREMELY restrictive rate limits on the free tier:

- User lookup: **3 requests per 15 minutes**
- Followers list: **Very limited** (exact limit unclear but very low)

This means your tool is practically unusable for real-time analysis without paying for higher tiers.

## Survival Strategies

### 1. Use Caching Aggressively

The updated code now caches:

- User data for 1 hour
- Followers lists for 30 minutes
- Run the same command twice - second time will be instant!

### 2. Batch Collection Mode

Instead of real-time analysis:

```bash
# Create a file with usernames (one per line)
echo -e "elonmusk\noprah\nbillgates" > targets.txt

# Collect data slowly (respects rate limits)
xknife batch collect-users targets.txt

# Analyze offline (no API calls!)
xknife batch analyze-file users_data_20230927_143022.json
```

### 3. Smart Usage Patterns

**✅ DO:**

- Cache everything locally
- Analyze the same users multiple times (uses cache)
- Use batch mode for multiple users
- Focus on quality over quantity

**❌ DON'T:**

- Try to analyze 100s of users in real-time
- Use followers command frequently (very expensive)
- Ignore rate limit warnings

### 4. Rate Limit Monitoring

The tool now shows:

- `[CACHE HIT]` - No API call made
- `[API CALL]` - Using your precious quota
- Rate limit warnings

### 5. Alternative Approaches

**For Production Use:**

1. **Upgrade to Twitter API Pro** ($100/month) - gets you 10,000 requests/month
2. **Use Academic Research Track** (free but requires approval)
3. **Focus on publicly available data** first
4. **Implement user-supplied data** workflows

**For Development:**

1. Use the new caching system extensively
2. Work with small datasets
3. Use mock data for testing
4. Focus on algorithm improvements over data collection

## Example Workflow

```bash
# 1. Clear cache to start fresh
xknife batch clear-cache

# 2. Analyze a single user (uses 1 API call)
xknife get --user someuser

# 3. Analyze same user again (uses cache, 0 API calls)
xknife get --user someuser

# 4. Create a target list
echo "user1" > my-targets.txt
echo "user2" >> my-targets.txt

# 5. Collect slowly over time
xknife batch collect-users my-targets.txt

# 6. Analyze offline as much as you want
xknife batch analyze-file users_data_*.json
```

## Command Reference

### New Commands

- `xknife batch collect-users <file>` - Slow, rate-limit-aware collection
- `xknife batch analyze-file <json>` - Fast offline analysis
- `xknife batch clear-cache` - Reset local cache

### Cache Locations

- Linux/Mac: `~/.xknife/cache/`
- Windows: `%USERPROFILE%\.xknife\cache\`

## Pro Tips

1. **Start Small**: Test with 2-3 users first
2. **Cache is King**: Run analysis multiple times on cached data
3. **Batch Everything**: Use collect-users for any multi-user analysis
4. **Monitor Carefully**: Watch for `[API CALL]` warnings
5. **Plan Ahead**: Each API call is precious, plan your research

Remember: The free tier is designed for experimentation, not production use!
